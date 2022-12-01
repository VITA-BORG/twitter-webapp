package main

import (
	"fmt"
	"time"

	"github.com/rainbowriverrr/F3Ytwitter/internal/models"
)

//ProfileWorker is the worker that scrapes the twitter profile of a user and stores it in the database concurrently.
//Then sends the user id to the follower or following channel if it falls below the limit
func (app *application) ProfileWorker() {
	//reads from the profile channel until it is closed
	for curr := range app.profileChan {
		currTime := time.Now()
		app.profileStatus = fmt.Sprintf("scraping %s", curr.Username)
		//always scrapes user because there will be updates
		user, err := app.scrapeUser(curr.Username)
		//checks if user is a participant.  If they are, it adds the relation with the school if it doesn't exist already
		if curr.IsParticipant {
			//add logic to create relation with school if it doesn't exist
			user.IsParticipant = true
		} else {
			user.IsParticipant = false
		}

		if err != nil {
			app.errorLog.Println("Error scraping user:", err)
			continue
		}
		//checks if the user is already in the database, if not, it adds it.
		if models.UserExists(app.connection, user.Handle) {
			app.infoLog.Println("User already exists in database")
			app.infoLog.Println("Updating user in database")
			err = models.UpdateUser(app.connection, user)
			if err != nil {
				app.errorLog.Println("Error updating user in database")
				app.errorLog.Println(err)
				app.profileStatus = "idle"
				continue
			}
		} else {
			app.infoLog.Println("User not in database")
			app.infoLog.Println("Adding user to database")
			err = models.InsertUser(app.connection, user)
			if err != nil {
				app.errorLog.Println("Error adding user to database")
				app.errorLog.Println(err)
				app.profileStatus = "idle"
				continue
			}
		}

		//checks if user is participant, and adds them to the students table if they are
		if curr.IsParticipant {
			//checks if student exists, adds them to the database if they don't
			if !models.StudentExists(app.connection, user.ID) {
				student := &models.Student{
					UserID:   user.ID,
					SchoolID: curr.ParticipantSchoolID,
					Cohort:   curr.ParticipantCohort,
				}
				err = models.InsertStudent(app.connection, student)
				if err != nil {
					app.errorLog.Println("Error adding student to database")
					app.errorLog.Println(err)
					app.profileStatus = "idle"
					continue
				}
			}
		}

		//checks if the user is a school.  If they are, it adds them to the school database including user database.
		if curr.IsSchool {
			app.infoLog.Println("User is a school")
			err = app.addSchool(curr.SchoolInfo)
			if err != nil {
				app.errorLog.Println("Error adding school to database")
				app.errorLog.Println(err)
				app.profileStatus = "idle"
				continue
			}
		}

		//adds biotags to the database
		tags := getBioTags(user.Bio)
		for _, tag := range tags {
			//checks if the tagged user is already in the database, if not, it adds it.
			if !models.UserExists(app.connection, tag) {
				taggedUser, err := app.scrapeUser(tag)
				if err != nil {
					app.errorLog.Println("Error scraping tagged user:", err)
					continue
				}
				err = models.InsertUser(app.connection, taggedUser)
				if err != nil {
					app.errorLog.Println("Error adding tagged user to database")
					app.errorLog.Println(err)
					continue
				}

				toAdd := &models.BioTag{
					UserID:          user.ID,
					MentionedUserID: taggedUser.ID,
					CollectedAt:     &currTime,
				}

				app.addBioTag(toAdd)
			} else {
				taggedUserID, err := models.GetUserIDByHandle(app.connection, tag)
				if err != nil {
					app.errorLog.Println("Error getting tagged user id:", err)
					continue
				}
				toAdd := &models.BioTag{
					UserID:          user.ID,
					MentionedUserID: taggedUserID,
					CollectedAt:     &currTime,
				}

				app.addBioTag(toAdd)
			}
		}

		//adds uid to simplifiedUser struct
		curr.ID = user.ID
		//Sends the simplifiedUser struct to the tweets channel.
		if curr.IsParticipant && curr.ScrapeContent {
			app.tweetsChan <- curr
		}

		//checks if user followers exceeds limit, if so, it does not scrape the followers
		if user.Followers > app.followLimit {
			app.infoLog.Println("User has too many followers, not scraping followers")
			app.profileStatus = "idle"
		} else if curr.ScrapeConnections {
			app.infoLog.Printf("Sending %s to followers channel", user.Handle)
			app.followerChan <- curr
		}

		//checks if user following exceeds limit, if so, it does not scrape the following
		if user.Following > app.followLimit {
			app.infoLog.Println("User has too many following, not scraping following")
			app.profileStatus = "idle"
		} else if curr.ScrapeConnections {
			app.infoLog.Printf("Sending %s to following channel", user.Handle)
			app.followChan <- curr
		}

		app.profileStatus = "idle"

	}
	app.profileStatus = "off"
	app.infoLog.Println("Profile Worker finished")
}

//TweetsWorker is the worker that scrapes the tweets of a user and stores them in the database concurrently
func (app *application) TweetsWorker() {
	//reads from the tweets channel until it is closed
	for user := range app.tweetsChan {
		app.tweetsStatus = fmt.Sprintf("scraping %s", user.Username)
		//scrapes tweets and updates them in database (includes retweets and replies)
		app.infoLog.Println("Scraping tweets for user:", user.ID)
		tweets := app.scrapeTweets(user.Username, user.StartDate)
		err := app.updateTweets(tweets)
		if err != nil {
			app.errorLog.Println("Error scraping tweets:", err)
			continue
		}
		//scrapes mentions and updates them in database
		app.infoLog.Println("Scraping mentions for user:", user.ID)
		userSlice, mentionsSlice := app.scrapeMentions(tweets)
		//double checks if the user is already in the database, if not, it adds it.
		for _, user := range userSlice {
			if !models.UserExists(app.connection, user.Handle) {
				err = models.InsertUser(app.connection, user)
				if err != nil {
					app.errorLog.Println("Error adding user to database")
					app.errorLog.Println(err)
					continue
				}
			}
		}
		//double checks if the mention is already in the database, if not, it adds it.
		for _, mention := range mentionsSlice {
			if !models.MentionExists(app.connection, mention) {
				err = models.InsertMention(app.connection, mention)
				if err != nil {
					app.errorLog.Println("Error adding mention to database")
					app.errorLog.Println(err)
					continue
				}
			}
		}
		app.tweetsStatus = "idle"
	}
	app.tweetsStatus = "off"
	app.infoLog.Println("Tweets Worker finished")
}

//Follow Worker is the worker that scrapes the followings of a user and stores them in the database concurrently.
func (app *application) FollowWorker() {
	//reads from the follow channel until it is closed
	for user := range app.followChan {
		app.followingStatus = fmt.Sprintf("scraping %s", user.Username)

		app.infoLog.Println("Scraping followings for user:", user.ID)

		upstreamChan := make(chan []*models.Follow)

		request := &followRequest{
			User:     user,
			upstream: upstreamChan,
		}
		//sends request for follows to the follow queue channel
		app.infoLog.Println("Sending request for follows for user:", user.ID)
		app.followQueue <- request
		//waits for the response from the follow queue channel
		app.infoLog.Printf("Waiting for follow response for user: %s", user.Username)
		follows := <-upstreamChan
		close(upstreamChan)
		if follows == nil {
			app.errorLog.Println("No follows found for user:", user.Username)
			continue
		}

		app.infoLog.Printf("%d follows recieved for user: %s", len(follows), user.Username)
		app.infoLog.Printf("Updating Follows for user: %s", user.Username)

		err := app.updateFollows(follows)
		if err != nil {
			app.errorLog.Println("Error updating followings:", err)
			continue
		}
		app.infoLog.Println("Sending request to connections worker for user:", user.Username)

		app.connectionsChan <- connectionsRequest{
			follows: follows,
			users:   "followings",
		}

		app.followStatus = "idle"
	}

	app.followingStatus = "off"
	app.infoLog.Println("Followings Worker finished")
}

//Follower Worker is the worker that scrapes the followers of a user and stores them in the database concurrently.
func (app *application) FollowerWorker() {
	//reads from the follower channel until it is closed
	for user := range app.followerChan {
		app.followStatus = fmt.Sprintf("scraping %s", user.Username)

		app.infoLog.Println("Scraping followers for user:", user.ID)

		upstreamChan := make(chan []*models.Follow)

		request := &followRequest{
			User:     user,
			upstream: upstreamChan,
		}

		//sends request for followers to the follower queue
		app.infoLog.Println("Sending request for follows for user:", user.ID)
		app.followerQueue <- request
		//waits for the response from the follower queue
		app.infoLog.Printf("Waiting for follower response for user: %s", user.Username)
		followers := <-upstreamChan
		close(upstreamChan)
		if followers == nil {
			app.errorLog.Println("No followers found for user:", user.Username)
			continue
		}
		app.infoLog.Printf("%d followers recieved for user: %s", len(followers), user.Username)
		app.infoLog.Printf("Updating Followers for user: %s", user.Username)
		err := app.updateFollows(followers)
		if err != nil {
			app.errorLog.Println("Error updating followers:", err)
			continue
		}

		app.infoLog.Println("Sending request to connections worker for user:", user.Username)
		app.connectionsChan <- connectionsRequest{
			follows: followers,
			users:   "followers",
		}

		app.followStatus = "idle"
	}
	//add connection scrape
	app.followStatus = "off"
	app.infoLog.Println("Follower Worker finished")
}

//Connections Worker is the worker that scrapes connections.  It acts as a queue for both the Follower Worker and Following worker in order to avoid rate limiting
func (app *application) ConnectionsWorker() {

	for request := range app.connectionsChan {

		var currentUser *simplifiedUser

		for _, user := range request.follows {
			//if the slice of follows is of followings of a user, that means the user is the followee, then that means the followerID and followerUsername is of the the other users.
			if request.users == "followings" {
				currentUser = &simplifiedUser{
					ID:       user.FolloweeID,
					Username: user.FolloweeUsername,
				}
			} else if request.users == "followers" {
				currentUser = &simplifiedUser{
					ID:       user.FollowerID,
					Username: user.FollowerUsername,
				}
			} else {
				app.errorLog.Println("Invalid user type")
				continue
			}

			//scrapes the user so that you can check for their follower and following count
			currUser, err := app.scrapeUser(currentUser.Username)
			if err != nil {
				app.errorLog.Println("Error scraping user:", err)
				continue
			}

			var followers []*models.Follow
			var followings []*models.Follow

			//sends a request to the followers and followings of the currentUser if they fall below the limit
			app.infoLog.Println("Sending request for followers for user:", currentUser.Username)
			if currUser.Followers < app.followLimit {
				followerChan := make(chan []*models.Follow)
				followerRequest := &followRequest{
					User:     currentUser,
					upstream: followerChan,
				}
				app.followerQueue <- followerRequest
				followers = <-followerChan
				close(followerChan)
				app.infoLog.Printf("%d followers recieved for user: %s", len(followers), currentUser.Username)
			}

			// TODO: Error with following scrape, always scraping the same person

			app.infoLog.Println("Sending request for followings for user:", currentUser.Username)
			if currUser.Following < app.followLimit {
				followingChan := make(chan []*models.Follow)
				followingRequest := &followRequest{
					User:     currentUser,
					upstream: followingChan,
				}
				app.followQueue <- followingRequest
				followings = <-followingChan
				close(followingChan)
				app.infoLog.Printf("%d followings recieved for user: %s", len(followings), currentUser.Username)
			}

			app.infoLog.Printf("Updating Connections for user: %s", currentUser.Username)
			//iterates through all followers and checks if they are already in the database
			//if they are already in the database, the follow is added since both users are already in the database
			for _, follower := range followers {
				if models.UserIDExists(app.connection, follower.FollowerID) {
					app.infoLog.Printf("Connection found. Follower: %s, Followee: %s", follower.FollowerUsername, follower.FolloweeUsername)
					models.InsertFollow(app.connection, follower)
				}
			}

			//iterates through all followings and checks if they are already in the database
			//if they are already in the database, the follow is added since both users are already in the database
			for _, following := range followings {
				if models.UserIDExists(app.connection, following.FolloweeID) {
					app.infoLog.Printf("Connection found. Follower: %s, Followee: %s", following.FollowerUsername, following.FolloweeUsername)
					models.InsertFollow(app.connection, following)
				}
			}

		}

	}

}

//FollowerQueue is a queue that reads from the follower channel for request structs which contain the channel where a slice of pointers to models.follow structs are returned.
func (app *application) FollowerQueue() {
	//reads from follower request channel and scrapes the requests
	for currRequest := range app.followerQueue {

		app.infoLog.Println("Recieved request for followers for user:", currRequest.User.Username)

		followers, err := app.getFollowers(currRequest.User)
		if err != nil {
			app.errorLog.Println("Error getting followers for: ", currRequest.User.Username)
			continue
		}

		currRequest.upstream <- followers

		//sleeps for a minute to avoid rate limiting
		time.Sleep(time.Minute)

	}
}

//FollowingQueue is a queue that reads from the following channel for request structs which contain the channel where a slice of pointers to models.folow structs are returned.
func (app *application) FollowingQueue() {
	//reads from following request channel and scrapes the requests
	for currRequest := range app.followQueue {

		app.infoLog.Println("Recieved request for followings for user:", currRequest.User.Username)

		followings, err := app.getFollows(currRequest.User)
		if err != nil {
			app.errorLog.Println("Error getting followings for: ", currRequest.User.Username)
			continue
		}

		currRequest.upstream <- followings

		//sleeps for a minute to avoid rate limiting
		time.Sleep(time.Minute)

	}
}
