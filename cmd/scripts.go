package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	twitterscraper "github.com/n0madic/twitter-scraper"
	"github.com/rainbowriverrr/F3Ytwitter/internal/models"
)

var keywords []string = []string{"official", "school", "institution", "program", "project", "institute",
	"faculty", "company", "team", "center", "conference", "organization", "we"}

//TODO: edit this to use goroutines and channels for different parts of the scrape
//Every table could have a go routine that takes information through channels

func resetTables(app *application) error {
	err := models.DeleteTables(app.connection)
	if err != nil {
		return err
	}
	err = models.CreateTables(app.connection)
	if err != nil {
		return err
	}
	return nil
}

//scrapeUser scrapes a user's twitter profile and returns a models.User struct.
//TODO: add error checking for handles that don't exist
func scrapeUser(app *application, handle string) (*models.User, error) {
	profile, err := app.scraper.GetProfile(handle)
	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}

	uid, err := strconv.ParseInt(profile.UserID, 10, 64)
	if err != nil {
		app.errorLog.Println(err)
		return nil, err
	}
	currTime := time.Now()

	fmt.Printf("%+v\n", profile)

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
	}, nil

}

//ScrapeTweets scrapes a user's tweets and returns a slice of models.Tweet structs.
//TODO: add scrape for tweets that are replies to other tweets (replies table)
//TODO: add scrape for users that are mentioned in tweets (mentions table)
func scrapeTweets(app *application, handle string, from time.Time) []*models.Tweet {
	cursor := ""
	var tweets []*twitterscraper.Tweet
	var err error
	var tweetsSlice []*models.Tweet //Slice to return
	numTweets := 0
	tweets, cursor, err = app.scraper.FetchTweets(handle, 200, cursor)
	if err != nil {
		app.errorLog.Println(err)
	}
	for tweets != nil {
		currTime := time.Now()
		for _, tweet := range tweets {

			//Checks if tweet is older than from date.  If it is, all remaining tweets are skipped.
			//scrapeTweets returns.
			if tweet.TimeParsed.Before(from) {
				app.infoLog.Printf("%d tweets scraped from %s", numTweets, handle)
				return tweetsSlice
			}

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
			var conversationID int64 = tweetID
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
			numTweets++
		}
		//Fetches next page of tweets
		tweets, cursor, err = app.scraper.FetchTweets(handle, 200, cursor)
		if err != nil {
			app.errorLog.Println(err)
		}
	}
	//returns slice of models.tweet structs if finished scraping and from date is never reached.
	app.infoLog.Printf("%d tweets scraped from %s", numTweets, handle)
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

//getMentions returns a slice of strings of the handles of users mentioned in a tweet.
func getMentions(bio string) []string {
	mentions := []string{}
	words := strings.Split(strings.ToLower(bio), " ")
	for _, curr := range words {
		if strings.HasPrefix(curr, "@") {
			curr = strings.TrimPrefix(curr, "@")
			curr = strings.Replace(curr, ",", "", 1)
			curr = strings.Replace(curr, ".", "", 1)
			curr = strings.Replace(curr, "!", "", 1)
			curr = strings.Replace(curr, "?", "", 1)
			curr = strings.Replace(curr, ":", "", 1)
			curr = strings.Replace(curr, ";", "", 1)
			curr = strings.Replace(curr, "…", "", 1)
			mentions = append(mentions, curr)
		}
	}
	return mentions
}

//scrapeMentions scrapes the users mentioned in a tweet, adds them to users table, and adds the tweet to the mentions table.
//TODO Parralelize
func scrapeMentions(app *application, tweets []*models.Tweet) error {

	re := regexp.MustCompile("@")

	for _, tweet := range tweets {
		if re.MatchString(tweet.Text) {
			var currUser *models.User
			var err error
			mentions := getMentions(tweet.Text)
			for _, mention := range mentions {
				app.infoLog.Printf("Scraped mention %s", mention)
				//checks to make sure user doesn't already exist before adding
				if !models.UserExists(app.connection, mention) {
					currUser, err = scrapeUser(app, mention)
					if err != nil {
						app.errorLog.Println("Error scraping user: ", err)
						app.errorLog.Println(err)
					}
					err = models.InsertUser(app.connection, *currUser)
					if err != nil {
						app.errorLog.Println(err)
					}
				}

				//checks to make sure User was successfully scraped before adding to mentions table
				if currUser != nil {
					//Creates models.Mention struct
					toInsert := models.Mention{
						UserID:  currUser.ID,
						TweetID: tweet.ID,
					}
					//checks to make sure mention doesn't already exist before adding
					if !models.MentionExists(app.connection, toInsert) {
						err = models.InsertMention(app.connection, toInsert)
						if err != nil {
							app.errorLog.Println(err)
						}
					}
				}
			}
		}
	}
	return nil
}
