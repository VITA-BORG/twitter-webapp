package models

import (
	"context"
	"time"

	pgx "github.com/jackc/pgx/v4"
)

type Tweet struct {
	ID             int64      `json:"id"`
	ConversationID int64      `json:"conversation_id"`
	Text           string     `json:"text"`
	PostedAt       *time.Time `json:"posted_at"`
	Url            string     `json:"url"`
	UserID         int64      `json:"user_id"`
	IsRetweet      bool       `json:"is_retweet"`
	RetweetID      int64      `json:"retweet_id"`
	Likes          int        `json:"likes"`
	Retweets       int        `json:"retweets"`
	Replies        int        `json:"replies"`
	CollectedAt    *time.Time `json:"collected_at"`
}

//insertTweet inserts a Tweet object into the database.  No checking.
func insertTweet(conn *pgx.Conn, tweet Tweet) error {
	statement := "INSERT INTO tweets(id, conversation_id, text, posted_at, url, user_id, is_retweet, retweet_id, likes, retweets, replies, collected_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)"
	_, err := conn.Exec(context.Background(), statement, tweet.ID, tweet.ConversationID, tweet.Text, tweet.PostedAt.Format(format), tweet.Url, tweet.UserID, tweet.IsRetweet, tweet.RetweetID, tweet.Likes, tweet.Retweets, tweet.Replies, tweet.CollectedAt.Format(format))
	return err
}
