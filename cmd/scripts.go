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

// scrapeUser scrapes a user's twitter profile and returns a models.User struct.
// TODO: add error checking for handles that don't exist
func (app *application) scrapeUser(handle string) (*models.User, error) {
	app.infoLog.Printf("Scraping user %s", handle)
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

// addOrUpdateUser adds a user to the database if it doesn't already exist.
// If the user already exists, it updates the user's information.
func (app *application) addOrUpdateUser(user *models.User) error {
	if !models.UserExists(app.connection, user.Handle) { //inserts user if they don't exist in the database
		err := models.InsertUser(app.connection, user)
		if err != nil {
			return err
		}
		app.infoLog.Printf("Added user %s", user.Handle)
	}
	//TODO: update user information if they already exist
	return nil
}

// getUserByHandle returns a user from the database based on a handle
// This adds a layer between the api handler and the database models
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

// getAllUsernames wraps getUsernames and returns a slice of strings.
// Uses app.errorLog to log errors.
func (app *application) getAllUsernames() []string {
	usernames, err := models.GetAllUsernames(app.connection)
	if err != nil {
		app.errorLog.Println(err)
	}
	return usernames
}

// ScrapeTweets scrapes a user's tweets and returns a slice of twitterscraper.Tweet structs.
// Note: Some retweets may be shortened.
func (app *application) scrapeTweets(handle string, from time.Time) []*twitterscraper.Tweet {
	cursor := ""
	var tweets []*twitterscraper.Tweet
	var err error
	var tweetsSlice []*twitterscraper.Tweet //Slice to return
	numTweets := 0
	tweets, cursor, err = app.scraper.FetchTweets(handle, 200, cursor)
	if err != nil {
		app.errorLog.Println(err)
	}
	for tweets != nil {

		for _, tweet := range tweets {

			//Checks if tweet is older than from date.  If it is, all remaining tweets are skipped.
			//scrapeTweets returns.
			if tweet.TimeParsed.Before(from) {
				app.infoLog.Printf("%d tweets scraped from %s", numTweets, handle)
				return tweetsSlice
			}

			tweetsSlice = append(tweetsSlice, tweet)
			numTweets++

		}

		tweets, cursor, err = app.scraper.FetchTweets(handle, 200, cursor)
		if err != nil {
			app.errorLog.Println(err)
		}
	}
	//returns slice of models.tweet structs if finished scraping and from date is never reached.
	app.infoLog.Printf("%d tweets scraped from %s", numTweets, handle)
	return tweetsSlice
}

// guessGender guesses the user's gender by looking for personal pronouns.
// If no personal pronouns are found, an empty string is returned
func guessGender(bio string) *string {

	var gender *string

	lowered := strings.ToLower(bio)
	matched, _ := regexp.MatchString(`they/them`, lowered)
	if matched {
		x := "X"
		gender = &x
		return gender
	}
	matched, _ = regexp.MatchString(`she/her`, lowered)
	if matched {
		f := "F"
		gender = &f
		return gender
	}
	matched, _ = regexp.MatchString(`he/him`, lowered)
	if matched {
		m := "M"
		gender = &m
		return gender
	}

	return gender
}

// isPerson checks if a user is a person by looking for keywords that indicate "non-person" status in their bio.
// If no keywords are found, true is returned.
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

// getMentions returns a slice of strings of the handles of users mentioned in a tweet.
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
			i := strings.Index(curr, "'")
			if i != -1 {
				curr = curr[:i]
			}
			mentions = append(mentions, curr)
		}
	}
	return mentions
}

// getBioTags scrapes the handles of users mentioned in a biography
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

// scrapeMentions scrapes the users mentioned in a tweet
// returns a slice of models.user structs of users mentioned in the tweet that do not exist in the database
// and a slice of models.mention structs of mentions that do not exist in the database
// Skips mentions in retweets by default.
// TODO Parralelize
func (app *application) scrapeMentions(tweets []*twitterscraper.Tweet, scrapeRetweets ...bool) ([]*models.User, []*models.Mention) {

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
			tweetID, err := strconv.ParseInt(tweet.ID, 10, 64)
			if err != nil {
				app.errorLog.Println(err)
				//mention is skipped if there is an issue with the tweetID
				continue
			}
			mentions := getMentions(tweet.Text)
			for _, mention := range mentions {
				app.infoLog.Printf("Scraped mention %s", mention)
				//checks to make sure user doesn't already exist before adding
				if !models.UserExists(app.connection, mention) {
					currUser, err = app.scrapeUser(mention)
					if err != nil {
						app.errorLog.Println("Error scraping user: ", err)
					} else {
						//Only user mention if user was successfully scraped
						userSlice = append(userSlice, currUser)
						toInsert := models.Mention{
							UserID:  currUser.ID,
							TweetID: tweetID,
						}
						//Only adds mention if user was successfully scraped
						//checks to make sure mention doesn't already exist before adding
						if !models.MentionExists(app.connection, &toInsert) {
							mentionSlice = append(mentionSlice, &toInsert)
						}
					}
				}
			}
		}
	}
	return userSlice, mentionSlice
}

// add Reply transforms a twitterscraper.Tweet to a models.Reply and adds it to the database
func (app *application) addReply(tweet *twitterscraper.Tweet) error {
	tweetID, err := strconv.ParseInt(tweet.ID, 10, 64)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}
	userRepliedToID, err := strconv.ParseInt(tweet.InReplyToStatus.UserID, 10, 64)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}
	//checks if userRepliedToID is in the database. If not, it is scraped.
	if !models.UserIDExists(app.connection, userRepliedToID) {
		userToAdd, err := app.scrapeUser(tweet.InReplyToStatus.Username)
		if err != nil {
			app.errorLog.Println("addReply: Error scraping user: ", err)
			return err
		}
		err = models.InsertUser(app.connection, userToAdd)
		if err != nil {
			app.errorLog.Println("addReply: Error inserting user: ", err)
			return err
		}
	}
	toAdd := models.Reply{
		TweetID: tweetID,
		ReplyID: userRepliedToID,
	}
	return models.InsertReply(app.connection, &toAdd)
}

// updateReplies checks if the reply exists in the database before adding it
// also adds user replied to if they do not exist.
func (app *application) updateReplies(replies []*models.Reply) error {
	for _, reply := range replies {
		if !models.ReplyExists(app.connection, reply) {
			err := models.InsertReply(app.connection, reply)
			if err != nil {
				app.errorLog.Println(err)
				return err
			}
		}
	}
	return nil
}

// addTweet transforms a twitterscraper.Tweet object into a models.Tweet object and adds it to the database
// checks if the tweet is a retweet, if it is, the retweet is added to the database if it does not already exist
// also adds hashtags
func (app *application) addTweet(tweet *twitterscraper.Tweet) error {

	now := time.Now()

	tweetID, err := strconv.ParseInt(tweet.ID, 10, 64)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	//does not add tweet if it already exists in database
	if models.TweetExists(app.connection, tweetID) {
		return nil
	}

	tweetUserID, err := strconv.ParseInt(tweet.UserID, 10, 64)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	//checks if user is in the database. If not, it is scraped.
	if !models.UserIDExists(app.connection, tweetUserID) {
		userToAdd, err := app.scrapeUser(tweet.Username)
		if err != nil {
			app.errorLog.Println("addTweet: Error scraping user: ", err)
			return err
		}
		err = models.InsertUser(app.connection, userToAdd)
		if err != nil {
			app.errorLog.Println(err)
			return err
		}
	}

	var conversationID int64 = tweetID //default is tweetID
	if tweet.IsReply {                 //todo: add reply to database if it does not already exist
		if tweet.InReplyToStatus != nil { //Extra check to make sure there actually is a tweet object
			app.addTweet(tweet.InReplyToStatus)
			app.addReply(tweet)
			conversationID, err = strconv.ParseInt(tweet.InReplyToStatus.ID, 10, 64)
			if err != nil {
				app.errorLog.Println(err)
				return err
			}
		}
	}

	var retweetID *int64
	if tweet.IsRetweet { //todo add retweet to database if it does not already exist
		if tweet.RetweetedStatus != nil { //Extra check to make sure there actually is a tweet object
			app.addTweet(tweet.RetweetedStatus)
			retweetIDint, err := strconv.ParseInt(tweet.RetweetedStatus.ID, 10, 64)
			if err != nil {
				app.errorLog.Println(err)
				return err
			}
			retweetID = &retweetIDint
		}
	} else if tweet.IsQuoted {
		if tweet.QuotedStatus != nil { //Extra check to make sure there actually is a tweet object
			app.addTweet(tweet.QuotedStatus)
			retweetIDint, err := strconv.ParseInt(tweet.QuotedStatus.ID, 10, 64)
			if err != nil {
				app.errorLog.Println(err)
				return err
			}
			retweetID = &retweetIDint
		}
	}

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
		CollectedAt:    &now,
	}

	//Adds tweet to database
	err = models.InsertTweet(app.connection, toAdd)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	//must insert tweet before inserting hashtag because it references tweet.ID
	//Adds hashtags to database
	for _, hashtag := range tweet.Hashtags {
		hashtagToAdd := &models.Hashtag{
			Hashtag: hashtag,
			TweetID: tweetID,
		}
		if !models.HashtagExists(app.connection, hashtagToAdd) {

			err = models.InsertHashtag(app.connection, hashtagToAdd)
			if err != nil {
				app.errorLog.Println(err)
				return err
			}
		}
	}

	return nil
}

// updateTweets updates the database with new tweets
func (app *application) updateTweets(tweets []*twitterscraper.Tweet) error {

	for _, tweet := range tweets {

		app.addTweet(tweet)

	}
	//Fetches
	return nil
}

// updateFollows updates the database with the new follows
// also updates the database with the new users
func (app *application) updateFollows(follows []*models.Follow) error {
	for _, follow := range follows {
		//check if the follow already exists in the database
		if !models.FollowExists(app.connection, follow) {
			//checks if the Followee exists in the database
			if !models.UserIDExists(app.connection, follow.FolloweeID) {
				//scrapes the user if it doesn't exist in the database
				user, err := app.scrapeUser(follow.FolloweeUsername)
				if err != nil {
					app.errorLog.Println("Error scraping user: ", err)
					app.errorLog.Println(err)
				}
				//adds the user to the database
				err = models.InsertUser(app.connection, user)
				if err != nil {
					app.errorLog.Println("Error inserting user: ", err)
					app.errorLog.Println(err)
				}
			}
			//checks if the Follower exists in the database
			if !models.UserIDExists(app.connection, follow.FollowerID) {
				//scrapes the user if it doesn't exist in the database
				user, err := app.scrapeUser(follow.FollowerUsername)
				if err != nil {
					app.errorLog.Println("Error scraping user: ", err)
					app.errorLog.Println(err)
				}
				//adds the user to the database
				err = models.InsertUser(app.connection, user)
				if err != nil {
					app.errorLog.Println("Error inserting user: ", err)
					app.errorLog.Println(err)
				}
			}

			err := models.InsertFollow(app.connection, follow)
			if err != nil {
				app.errorLog.Println(err)
				return err
			}
		}
	}
	return nil
}

// addSchool adds a school in the database.  It will also assign the school an ID.
func (app *application) addSchool(school *simplifiedSchool) error {
	var toAdd models.School

	user, err := app.scrapeUser(school.TwitterHandle)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	err = app.addOrUpdateUser(user)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	currNum, err := models.NumberOfSchools(app.connection)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	toAdd.ID = currNum + 1
	toAdd.Name = school.Name
	toAdd.TopRated = school.TopRated
	toAdd.Public = school.Public
	toAdd.City = school.City
	toAdd.State = school.State
	toAdd.Country = school.Country
	toAdd.User_ID = user.ID

	err = models.InsertSchool(app.connection, &toAdd)
	if err != nil {
		app.errorLog.Println(err)
		return err
	}

	return nil
}

// addBioTag adds a biotag to the database, checks if it already exists
func (app *application) addBioTag(bioTag *models.BioTag) error {
	if !models.TagExists(app.connection, bioTag.UserID, bioTag.MentionedUserID) {
		err := models.InsertBioTag(app.connection, bioTag)
		if err != nil {
			app.errorLog.Println(err)
			return err
		}
	}
	return nil
}

// converts the simplifiedUser to a simple request.  All unnecessary fields are removed
func simpleUsertoSimpleRequest(user *simplifiedUser) *models.SimpleRequest {
	return &models.SimpleRequest{
		ID:                 user.BackupID,
		UID:                user.ID,
		Username:           user.Username,
		Scrape_connections: user.ScrapeConnections,
	}
}

// converts the simple request to a simplified user.
func simpleRequstToSimpleUser(request *models.SimpleRequest) *simplifiedUser {
	return &simplifiedUser{
		ID:                request.UID,
		Username:          request.Username,
		ScrapeConnections: request.Scrape_connections,
		BackupID:          request.ID,
	}
}

func (app *application) populateFollowQueue(users []*models.SimpleRequest) {
	for _, user := range users {
		app.followChan <- user
	}
}

func (app *application) populateFollowerQueue(users []*models.SimpleRequest) {
	for _, user := range users {
		app.followerChan <- user
	}
}

// populateConnectionRequest takes a models.ConnectionRequest, queries the proper followers/followings and creates a connectionRequest struct
func (app *application) populateConnectionRequest(request *models.ConnectionRequest) *connectionsRequest {
	var follows []*models.Follow
	var err error
	populatedRequest := &connectionsRequest{
		ID:    request.ID,
		users: request.FollowsOrFollowers,
	}
	if request.FollowsOrFollowers == "follows" {
		populatedRequest.users = "followings"
		//query for follows
		follows, err = models.GetFollows(app.connection, request.UID)
		if err != nil {
			app.errorLog.Println("Error: " + err.Error())
			return nil
		}
	} else if request.FollowsOrFollowers == "followers" {
		populatedRequest.users = "followers"
		//query for followers
		follows, err = models.GetFollowers(app.connection, request.UID)
		if err != nil {
			app.errorLog.Println("Error: " + err.Error())
			return nil
		}
	}

	populatedRequest.follows = follows

	return populatedRequest
}

func (app *application) loadBackups() {
	//Query database for all follower_requests
	requests, err := models.GetSimpleRequests(app.connection, "followers")
	if err != nil {
		app.errorLog.Println("Error: " + err.Error())
		return
	}
	//check if there are any requests
	if len(requests) > 0 {
		//populate the follower queue
		app.populateFollowerQueue(requests)
	}

	//Query database for all follow_requests
	requests, err = models.GetSimpleRequests(app.connection, "follows")
	if err != nil {
		app.errorLog.Println("Error: " + err.Error())
		return
	}
	//check if there are any requests
	if len(requests) > 0 {
		//populate the follow queue
		app.populateFollowQueue(requests)
	}

	//Query database for all connection_requests
	connectionRequests, err := models.GetConnectionRequests(app.connection)
	if err != nil {
		app.errorLog.Println("Error: " + err.Error())
		return
	}
	//check if there are any requests
	if len(connectionRequests) == 0 {
		app.infoLog.Printf("no connection requests found")
		return
	}

	//populate the connection reqeusts
	for _, request := range connectionRequests {
		app.infoLog.Printf("Loading connection request: %v", request.ID)
		app.connectionsChan <- *app.populateConnectionRequest(request)
	}

}
