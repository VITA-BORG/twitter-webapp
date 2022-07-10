package main

import (
	"time"

	"github.com/rainbowriverrr/F3Ytwitter/internal/models"
)

//ProfileWorker is the worker that scrapes the twitter profile of a user and stores it in the database concurrently.
//Then sends the user id to the follower or following channel if it falls below the limit
func (app *application) ProfileWorker() {
	//reads from the profile channel until it is closed
	for handle := range app.profileChan {
		user, err := app.scrapeUser(handle)
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
				continue
			}
		} else {
			app.infoLog.Println("User not in database")
			app.infoLog.Println("Adding user to database")
			err = models.InsertUser(app.connection, user)
			if err != nil {
				app.errorLog.Println("Error adding user to database")
				app.errorLog.Println(err)
				continue
			}
		}
		//checks if user followers exceeds limit, if so, it does not scrape the followers
		if user.Followers > app.followLimit {
			app.infoLog.Println("User has too many followers, not scraping followers")
			continue
		}
		//sends uid to follower channel
		app.followerChan <- simplifiedUser{ID: user.ID, Username: user.Handle}
		//checks if user following exceeds limit, if so, it does not scrape the following
		if user.Following > app.followLimit {
			app.infoLog.Println("User has too many following, not scraping following")
			continue
		}
		//sends uid to follow channel
		app.followChan <- simplifiedUser{ID: user.ID, Username: user.Handle}

	}
	app.infoLog.Println("Profile Worker finished")
}

//Follow Worker is the worker that scrapes the followers of a user and stores them in the database concurrently.
func (app *application) FollowWorker() {
	//reads from the follow channel until it is closed
	for uid := range app.followChan {
		follows, err := app.getFollows(uid)
		if err != nil {
			app.errorLog.Println("Error getting follows:", err)
			continue
		}
		app.updateFollows(follows)
		time.Sleep(60 * time.Second) //sleep 60 seconds to avoid rate limiting just in case
	}
	app.infoLog.Println("Follow Worker finished")
}

//Follower Worker is the worker that scrapes the followers of a user and stores them in the database concurrently.
func (app *application) FollowerWorker() {
	//reads from the follower channel until it is closed
	for uid := range app.followerChan {
		followers, err := app.getFollowers(uid)
		if err != nil {
			app.errorLog.Println("Error getting followers:", err)
			continue
		}
		app.updateFollows(followers)
		time.Sleep(60 * time.Second) //sleep 60 seconds to avoid rate limiting just in case
	}
	app.infoLog.Println("Follower Worker finished")
}
