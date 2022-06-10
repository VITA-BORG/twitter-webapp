package main

import (
	"context"
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

	uid, err := strconv.ParseInt(profile.UserID, 10, 64)
	if err != nil {
		app.errorLog.Println(err)
	}
	currTime := time.Now()

	fmt.Printf("%+v\n", profile)

	//TODO: gender script, IsPerson script, time.Format
	return &models.User{
		ID:          uid,
		ProfileName: profile.Name,
		Handle:      profile.Username,
		Gender:      guessGender(profile.Biography),
		IsPerson:    isPerson(profile.Biography),
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

//ScrapeTweets scrapes a user's tweets and returns a slice of models.Tweet structs.
//TODO: add scrape for tweets that are replies to other tweets (replies table)
//TODOL : add scrape for users that are mentioned in tweets (mentions table)
//TODO: add ability to scrape starting from date
func scrapeTweets(app *application, handle string) []*models.Tweet {
	tweets := app.scraper.GetTweets(context.Background(), handle, 10)
	currTime := time.Now()
	curr := 0
	var tweetsSlice []*models.Tweet
	tweetsSlice = make([]*models.Tweet, len(tweets))
	for tweet := range tweets {

		//app.infoLog.Printf(tweet.TimeParsed.Format("2006-01-02"))

		//Converts fields to appropriate types for DB model
		tweetUserID, err := strconv.ParseInt(tweet.UserID, 10, 64)
		if err != nil {
			app.errorLog.Println(err)
		}
		tweetID, err := strconv.ParseInt(tweet.ID, 10, 64)
		if err != nil {
			app.errorLog.Println(err)
		}
		var conversationID int64 = 0
		if tweet.IsReply {
			if tweet.InReplyToStatus != nil { //Extra check to make sure there actually is a tweet object
				conversationID, err = strconv.ParseInt(tweet.InReplyToStatus.ID, 10, 64)
				if err != nil {
					app.errorLog.Println(err)
				}
			}
		}
		var retweetID int64 = 0
		if tweet.IsRetweet {
			if tweet.RetweetedStatus != nil { //Extra check to make sure there actually is a tweet object
				retweetID, err = strconv.ParseInt(tweet.RetweetedStatus.ID, 10, 64)
				if err != nil {
					app.errorLog.Println(err)
				}
			}
		} else if tweet.IsQuoted {
			if tweet.QuotedStatus != nil { //Extra check to make sure there actually is a tweet object
				retweetID, err = strconv.ParseInt(tweet.QuotedStatus.ID, 10, 64)
				if err != nil {
					app.errorLog.Println(err)
				}
			}
		}

		//Creates models.tweet struct
		toAdd := &models.Tweet{
			ID:             tweetID,
			ConversationID: conversationID,
			Text:           tweet.Text,
			PostedAt:       &tweet.TimeParsed,
			Url:            tweet.PermanentURL,
			UserID:         tweetUserID,
			IsRetweet:      tweet.IsRetweet,
			RetweetID:      retweetID,
			Likes:          tweet.Likes,
			Retweets:       tweet.Retweets,
			Replies:        tweet.Replies,
			CollectedAt:    &currTime,
		}
		tweetsSlice = append(tweetsSlice, toAdd)
		curr++
	}
	app.infoLog.Printf("%d tweets scraped for %s", curr, handle)
	return tweetsSlice
}

//guessGender guesses the user's gender by looking for personal pronouns.
//If no personal pronouns are found, an empty string is returned
func guessGender(bio string) string {
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

//isPerson checks if a user is a person by looking for keywords that indicate "non-person" status in their bio.
//If no keywords are found, true is returned.
func isPerson(bio string) bool {
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
