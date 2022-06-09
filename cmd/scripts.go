package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rainbowriverrr/F3Ytwitter/internal/models"
)

var keywords []string = []string{"official", "school", "institution", "program", "project", "institute",
	"faculty", "company", "team", "center", "conference", "organization", "we"}

//scrapeUser scrapes a user's twitter profile and returns a models.User struct.
func scrapeUser(app *application, handle string) *models.User {
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
	return &models.User{
		Id:          uid,
		ProfileName: profile.Name,
		Handle:      profile.Username,
		Gender:      guess_gender(profile.Biography),
		IsPerson:    is_person(profile.Biography),
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

//guess_gender guesses the user's gender by looking for personal pronouns.
//If no personal pronouns are found, an empty string is returned
func guess_gender(bio string) string {
	lowered := strings.ToLower(bio)
	matched, _ := regexp.MatchString(`/?they/?`, lowered)
	if matched {
		return "X"
	}
	matched, _ = regexp.MatchString(`/?she/?`, lowered)
	if matched {
		return "F"
	}
	matched, _ = regexp.MatchString(`/?he/?`, lowered)
	if matched {
		return "M"
	}

	return ""
}

//is_person checks if a user is a person by looking for keywords that indicate "non-person" status in their bio.
//If no keywords are found, true is returned.
func is_person(bio string) bool {
	words := strings.Split(strings.ToLower(bio), " ")
	for _, curr := range words {
		for _, keyword := range keywords {
			if keyword == curr {
				return false
			}
		}
	}
	return true
}
