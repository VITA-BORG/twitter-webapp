package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/rainbowriverrr/F3Ytwitter/internal/models"
)

//scrapeUser scrapes a user's twitter profile and returns a models.User struct.
func scrapeUser(app *application, handle string) models.User {
	profile, err := app.scraper.GetProfile(handle)
	if err != nil {
		app.errorLog.Println(err)
	}

	uid, err := strconv.Atoi(profile.UserID)
	if err != nil {
		app.errorLog.Println(err)
	}
	currTime := time.Now()

	fmt.Printf("%+v\n", profile)

	//TODO: gender script, IsPerson script, time.Format
	return models.User{
		Id:          uid,
		ProfileName: profile.Name,
		Handle:      profile.Username,
		Gender:      "X",
		IsPerson:    true,
		Joined:      profile.Joined,
		Bio:         profile.Biography,
		Location:    profile.Location,
		Verified:    profile.IsVerified,
		Avatar:      profile.Avatar,
		Tweets:      profile.TweetsCount,
		Likes:       profile.LikesCount,
		Media:       0,
		Following:   profile.FollowingCount,
		Followers:   profile.FollowersCount,
		CollectedAt: &currTime,
	}

}
