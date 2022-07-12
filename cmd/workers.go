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
		app.profileStatus = fmt.Sprintf("scraping %s", curr.Username)
		user, err := app.scrapeUser(curr.Username)
		if curr.isParticipant {
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
		//adds uid to simplifiedUser struct
		curr.ID = user.ID
		//Sends the simplifiedUser struct to the tweets channel.
		if curr.isParticipant {
			app.tweetsChan <- curr
		}

		//checks if user followers exceeds limit, if so, it does not scrape the followers
		if user.Followers > app.followLimit {
			app.infoLog.Println("User has too many followers, not scraping followers")
			app.profileStatus = "idle"
			continue
		}

		app.followerChan <- curr
		//checks if user following exceeds limit, if so, it does not scrape the following
		if user.Following > app.followLimit {
			app.infoLog.Println("User has too many following, not scraping following")
			app.profileStatus = "idle"
			continue
		}
		app.followChan <- curr
		app.profileStatus = "idle"

	}
	app.profileStatus = "off"
	app.infoLog.Println("Profile Worker finished")
}

//TweetsWorker is the worker that scrapes the tweets of a user and stores them in the database concurrently
func (app *application) TweetsWorker() {
	//reads from the tweets channel until it is closed
	for user := range app.tweetsChan {
		app.tweetsStatus = fmt.Sprintf("scraping %d", user.ID)
		//scrapes tweets and updates them in database
		app.infoLog.Println("Scraping tweets for user:", user.ID)
		tweets := app.scrapeTweets(user.Username, user.startDate)
		err := app.updateTweets(tweets)
		if err != nil {
			app.errorLog.Println("Error scraping tweets:", err)
			continue
		}
		//checks for replies and adds them to the database
		app.infoLog.Println("Sraping replies for user:", user.ID)
		replies, err := app.getReplies(tweets)
		if err != nil {
			app.errorLog.Println("Error scraping replies:", err)
			continue
		}
		err = app.updateReplies(replies)
		if err != nil {
			app.errorLog.Println("Error updating replies:", err)
			continue
		}
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
		app.followingStatus = fmt.Sprintf("scraping %d", user.ID)
		follows, err := app.getFollows(user)
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
		app.followStatus = "idle"
	}
	app.followingStatus = "off"
	app.infoLog.Println("Followings Worker finished")
}

//Follower Worker is the worker that scrapes the followers of a user and stores them in the database concurrently.
func (app *application) FollowerWorker() {
	//reads from the follower channel until it is closed
	for user := range app.followerChan {
		app.followStatus = fmt.Sprintf("scraping %d", user.ID)
		followers, err := app.getFollowers(user)
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
	app.followStatus = "off"
	app.infoLog.Println("Follower Worker finished")
}
