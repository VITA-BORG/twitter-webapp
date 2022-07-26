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
			app.followerChan <- curr
		}

		//checks if user following exceeds limit, if so, it does not scrape the following
		if user.Following > app.followLimit {
			app.infoLog.Println("User has too many following, not scraping following")
			app.profileStatus = "idle"
		} else if curr.ScrapeConnections {
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
		follows, err := app.getFollows(*user)
		if err != nil {
			app.errorLog.Println("Error getting followings:", err)
			continue
		}
		err = app.updateFollows(follows)
		if err != nil {
			app.errorLog.Println("Error updating followings:", err)
			continue
		}
		time.Sleep(60 * time.Second) //sleep 60 seconds to avoid rate limiting just in case
		//Connection scrape, sees if there are any connections between followings of the user and the rest in the database
		//app.infoLog.Println("Scraping connections for user:", user.Username)
		// for _, follow := range follows {
		// 	simpliefiedFollow := app.simplifyFollowing(follow)
		// 	followFollowings, err := app.getFollows(*simpliefiedFollow)
		// 	if err != nil {
		// 		app.errorLog.Println("Error getting followings:", err)
		// 		continue
		// 	}
		// 	//checks if the followings of the follow are already in the database,.  If they are, a follow is added to the database.  If they aren't they are skipped.
		// 	for _, followFollowing := range followFollowings {
		// 		if models.UserIDExists(app.connection, followFollowing.FolloweeID) {
		// 			//insert follow
		// 			err = models.InsertFollow(app.connection, followFollowing)
		// 			if err != nil {
		// 				app.errorLog.Println("Error inserting follow:", err)
		// 				continue
		// 			}
		// 		}
		// 	}

		// 	//sleeps for a minute to avoid rate limiting just in case
		// 	time.Sleep(60 * time.Second)

		// 	followFollowers, err := app.getFollowers(*simpliefiedFollow)
		// 	if err != nil {
		// 		app.errorLog.Println("Error getting followers:", err)
		// 		continue
		// 	}
		// 	//checks if the followers of the follow are already in the database,.  If they are, a follow is added to the database.  If they aren't they are skipped.
		// 	for _, followFollower := range followFollowers {
		// 		if models.UserIDExists(app.connection, followFollower.FollowerID) {
		// 			//insert follow
		// 			err = models.InsertFollow(app.connection, followFollower)
		// 			if err != nil {
		// 				app.errorLog.Println("Error inserting follow:", err)
		// 				continue
		// 			}
		// 		}
		// 	}

		// 	//sleeps for a minute to avoid rate limiting just in case
		// 	time.Sleep(60 * time.Second)

		//}
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
		followers, err := app.getFollowers(*user)
		if err != nil {
			app.errorLog.Println("Error getting followers:", err)
			continue
		}
		err = app.updateFollows(followers)
		if err != nil {
			app.errorLog.Println("Error updating followers:", err)
			continue
		}
		time.Sleep(60 * time.Second) //sleep 60 seconds to avoid rate limiting just in case
		app.followStatus = "idle"
	}
	//add connection scrape
	app.followStatus = "off"
	app.infoLog.Println("Follower Worker finished")
}
