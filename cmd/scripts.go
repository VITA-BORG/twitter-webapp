package main

import (
	"errors"
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

func (app *application) resetTables() error {
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
func (app *application) scrapeUser(handle string) (*models.User, error) {
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

	//app.infoLog.Printf("%+v\n", profile)

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
//Note: Some retweets may be shortened.
func (app *application) scrapeTweets(handle string, from time.Time) []*models.Tweet {
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
func guessGender(bio string) *string {

	var gender string

	lowered := strings.ToLower(bio)
	matched, _ := regexp.MatchString(`/?they/?`, lowered)
	if matched {
		gender = "X"
		return &gender
	}
	matched, _ = regexp.MatchString(`/?she/?`, lowered)
	if matched {
		gender = "F"
		return &gender
	}
	matched, _ = regexp.MatchString(`/?he/?`, lowered)
	if matched {
		gender = "M"
		return &gender
	}

	return &gender
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
func getMentions(text string) []string {
	mentions := []string{}
	words := strings.Split(strings.ToLower(text), " ")
	//removes the retweeted's handle from the list of mentions
	if len(words) > 1 {
		if words[0] == "rt" {
			words = words[2:]
		}
	}

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

//getBioTags scrapes the handles of users mentioned in a biography
func getBioTags(bio string) []string {
	tags := []string{}
	//removes all non-alphanumeric characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9_]`)
	words := strings.Split(strings.ToLower(bio), " ")
	for _, curr := range words {
		if strings.HasPrefix(curr, "@") {
			curr = reg.ReplaceAllString(curr, "")
			tags = append(tags, curr)
		}
	}
	return tags

}

//scrapeMentions scrapes the users mentioned in a tweet
//returns a slice of models.user structs of users mentioned in the tweet that do not exist in the databas
//and a slice of models.mention structs of mentions that do not exist in the database
//Skips mentions in retweets by default.
//TODO Parralelize
func (app *application) scrapeMentions(tweets []*models.Tweet, scrapeRetweets ...bool) ([]*models.User, []*models.Mention) {

	//By default, retweets are skipped
	scrapeRT := false
	if len(scrapeRetweets) > 0 {
		scrapeRT = scrapeRetweets[0]
	}

	var userSlice []*models.User
	var mentionSlice []*models.Mention

	re := regexp.MustCompile("@")

	for _, tweet := range tweets {
		if tweet.IsRetweet && !scrapeRT {
			continue
		}
		if re.MatchString(tweet.Text) {
			var currUser *models.User
			var err error
			mentions := getMentions(tweet.Text)
			for _, mention := range mentions {
				app.infoLog.Printf("Scraped mention %s", mention)
				//checks to make sure user doesn't already exist before adding
				if !models.UserExists(app.connection, mention) {
					currUser, err = app.scrapeUser(mention)
					if err != nil {
						app.errorLog.Println("Error scraping user: ", err)
						app.errorLog.Println(err)
					} else {
						//Only user mention if user was successfully scraped
						userSlice = append(userSlice, currUser)
						toInsert := models.Mention{
							UserID:  currUser.ID,
							TweetID: tweet.ID,
						}
						//Only adds mention if user was successfully scraped
						//checks to make sure mention doesn't already exist before adding
						if !models.MentionExists(app.connection, toInsert) {
							mentionSlice = append(mentionSlice, &toInsert)
						}
					}
				}
			}
		}
	}
	return userSlice, mentionSlice
}

//isReply checks if a tweet is a reply to another tweet
func isReply(tweet *models.Tweet) bool {
	return tweet.ID != tweet.ConversationID
}

//getReplies returns a slice of models.reply structs of tweets that are replies
func getReplies(tweets []*models.Tweet) []*models.Reply {
	var replySlice []*models.Reply
	for _, tweet := range tweets {
		if isReply(tweet) {
			toAdd := models.Reply{
				TweetID: tweet.ID,
				ReplyID: tweet.ConversationID,
			}
			replySlice = append(replySlice, &toAdd)
		}
	}
	return replySlice
}

func (app *application) getUserByHandle(handle string) (*models.User, error) {

	var user *models.User
	var err error

	if models.UserExists(app.connection, handle) {
		user, err = models.GetUserByHandle(app.connection, handle)
		app.infoLog.Printf("User %s fetched from database", handle)
		return user, err
	} else {
		app.errorLog.Printf("User %s does not exist in database", handle)
		return nil, errors.New("user does not exist")
	}

}
