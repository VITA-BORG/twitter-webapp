package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rainbowriverrr/F3Ytwitter/internal/models"
)

//singleFollow is the follower object that is expected from the twitter api (all strings)
type singleFollow struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	Name      string `json:"name"`
	Username  string `json:"username"`
}

//meta is the expected "meta" object in the response from the twitter api
type meta struct {
	NextToken     string `json:"next_token"`
	PreviousToken string `json:"previous_token"`
	ResultCount   int    `json:"result_count"`
}

//followerResponse is the struct of the expected response from the twitter api
type followerResponse struct {
	Data []singleFollow `json:"data"`
	Meta meta           `json:"meta"`
}

//getURL returns the twitter v2 api url for a given endpoint
func getFollowURL(uid int64, followStatus string, pageToken string) string {
	url := fmt.Sprintf("https://api.twitter.com/2/users/%d/%s?user.fields=created_at&max_results=1000", uid, followStatus)
	if pageToken != "" {
		url += "&pagination_token=" + pageToken
	}
	return url
}

//getResponse returns the response from a given url.  This is used in tandem with getURL to get the response from the twitter api
//This also adds important headers to the request
func (app *application) getResponse(url string) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+app.bearerToken)
	request.Header.Set("Content-Type", "application/json")

	resp, err := app.followerClient.Do(request)
	if err != nil {
		app.errorLog.Println("Error getting response from url:", url)
		return nil, err
	}

	return resp, nil

}

//getFollowers scrapes a user's profile and returns a slice of models.Follow structs of users that follow a user
func (app *application) getFollowers(uid int64) ([]*models.Follow, error) {
	var followers []*models.Follow
	var pageToken string
	currTime := time.Now()
	//get the first page of followers, then continues going as long as there is a next page
	for {
		url := getFollowURL(uid, "followers", pageToken)
		resp, err := app.getResponse(url)
		if err != nil {
			app.errorLog.Println(err)
			return nil, err
		}
		defer resp.Body.Close()

		//checks if the response is ok
		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				app.errorLog.Println(err)
				return nil, err
			}

			//unmarshal the response into a followerResponse struct
			var followerResponse followerResponse
			err = json.Unmarshal(body, &followerResponse)
			if err != nil {
				app.errorLog.Println(err)
				return nil, err
			}

			if app.debug {
				app.infoLog.Printf("%v\n", followerResponse)
			}

			//iterate through the followers and add them to the slice
			for _, follower := range followerResponse.Data {

				followerID, err := strconv.ParseInt(follower.ID, 10, 64)
				if err != nil {
					app.errorLog.Println("error converting follower id to int:", err)
					return nil, err
				}

				//trims the extra characters from the twitter response
				toTrim := strings.Index(follower.CreatedAt, "T")
				if toTrim == -1 {
					continue
				}
				follower.CreatedAt = follower.CreatedAt[:toTrim]
				createdAt, err := time.Parse(models.Format, follower.CreatedAt)
				if err != nil {
					app.errorLog.Println("error parsing time:", err)
					return nil, err
				}

				followers = append(followers, &models.Follow{
					FollowerID:  followerID,
					CreatedAt:   createdAt,
					FolloweeID:  uid,
					CollectedAt: currTime,
				})
			}
			//if there is a next page, set the page token to the next page
			//otherwise, break out of the loop
			pageToken = followerResponse.Meta.NextToken
			if pageToken == "" {
				break
			}
			//sleep 60 seconds to avoid rate limiting
			time.Sleep(60 * time.Second)
		} else {
			app.errorLog.Printf("error getting followers for user %d.  Status code: %d\n", uid, resp.StatusCode)
			return nil, fmt.Errorf("error getting followers for user %d.  Status code: %d", uid, resp.StatusCode)
		}

	}
	return followers, nil
}

//getFollows scrapes a user's profile and returns a slice of models.Follow structs of users that a user follows
func (app *application) getFollows(uid int64) ([]*models.Follow, error) {
	var follows []*models.Follow
	var pageToken string
	currTime := time.Now()
	//get the first page of followers, then continues going as long as there is a next page
	for {
		url := getFollowURL(uid, "following", pageToken)
		resp, err := app.getResponse(url)
		if err != nil {
			app.errorLog.Println(err)
			return nil, err
		}
		defer resp.Body.Close()

		//checks if the response is ok
		if resp.StatusCode == http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				app.errorLog.Println(err)
				return nil, err
			}

			//unmarshal the response into a followerResponse struct
			var followerResponse followerResponse
			err = json.Unmarshal(body, &followerResponse)
			if err != nil {
				app.errorLog.Println(err)
				return nil, err
			}

			if app.debug {
				app.infoLog.Printf("%v\n", followerResponse)
			}

			//iterate through the followers and add them to the slice
			for _, follower := range followerResponse.Data {

				followerID, err := strconv.ParseInt(follower.ID, 10, 64)
				if err != nil {
					app.errorLog.Println("error converting follower id to int:", err)
					return nil, err
				}

				//trims the extra characters from the twitter response
				toTrim := strings.Index(follower.CreatedAt, "T")
				if toTrim == -1 {
					continue
				}
				follower.CreatedAt = follower.CreatedAt[:toTrim]
				createdAt, err := time.Parse(models.Format, follower.CreatedAt)
				if err != nil {
					app.errorLog.Println("error parsing time:", err)
					return nil, err
				}

				follows = append(follows, &models.Follow{
					FollowerID:  uid,
					CreatedAt:   createdAt,
					FolloweeID:  followerID,
					CollectedAt: currTime,
				})
			}
			//if there is a next page, set the page token to the next page
			//otherwise, break out of the loop
			pageToken = followerResponse.Meta.NextToken
			if pageToken == "" {
				break
			}
			//sleep 60 seconds to avoid rate limiting
			time.Sleep(60 * time.Second)
		} else {
			app.errorLog.Printf("error getting followers for user %d.  Status code: %d\n", uid, resp.StatusCode)
			return nil, fmt.Errorf("error getting followers for user %d.  Status code: %d", uid, resp.StatusCode)
		}

	}
	return follows, nil
}
