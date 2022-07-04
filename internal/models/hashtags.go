package models

import (
	"context"

	pgx "github.com/jackc/pgx/v4"
)

type Hashtag struct {
	ID      int64  `json:"id"`
	TweetID int64  `json:"tweet_id"`
	Hashtag string `json:"hashtag"`
}

//InsertHashtag inserts a Hashtag object into the database.  No checking.
func InsertHashtag(conn *pgx.Conn, hashtag Hashtag) error {
	statement := "INSERT INTO hashtags(tag, tweet_id) VALUES($1, $2)"
	_, err := conn.Exec(context.Background(), statement, hashtag.Hashtag, hashtag.TweetID)
	return err
}
