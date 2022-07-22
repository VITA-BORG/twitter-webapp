package models

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Hashtag struct {
	ID      int64  `json:"id"`
	TweetID int64  `json:"tweet_id"`
	Hashtag string `json:"tag"`
}

//InsertHashtag inserts a Hashtag object into the database.  No checking.
func InsertHashtag(conn *pgxpool.Pool, hashtag *Hashtag) error {
	statement := "INSERT INTO hashtags(tag, tweet_id) VALUES($1, $2)"
	_, err := conn.Exec(context.Background(), statement, hashtag.Hashtag, hashtag.TweetID)
	return err
}

//HashtagExists checks if a hashtag exists in the database.
func HashtagExists(conn *pgxpool.Pool, hashtag *Hashtag) bool {
	var exists bool
	statement := "SELECT EXISTS(SELECT 1 FROM hashtags WHERE tag=$1 AND tweet_id=$2)"
	err := conn.QueryRow(context.Background(), statement, hashtag.Hashtag, hashtag.TweetID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
