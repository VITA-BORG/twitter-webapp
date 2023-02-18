package models

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Reply struct {
	ID      int64 `json:"id"`
	TweetID int64 `json:"tweet_id"`
	ReplyID int64 `json:"user_replied_to_id"`
}

// InsertReply inserts a Reply object into the database.  No checking.
func InsertReply(conn *pgxpool.Pool, reply *Reply) error {
	if ReplyExists(conn, reply) {
		return nil
	}
	statement := "INSERT INTO replies(tweet_id, user_id) VALUES($1, $2)"
	_, err := conn.Exec(context.Background(), statement, reply.TweetID, reply.ReplyID)
	return err
}

// GetReplyByID returns a Reply object from the database if they exist.  Otherwise, it returns nil.
func GetReplyByID(conn *pgxpool.Pool) (Reply, error) {
	var reply Reply
	var err error
	statement := "SELECT * FROM replies WHERE tweet_id=$1"
	err = conn.QueryRow(context.Background(), statement).Scan(&reply.TweetID, &reply.ReplyID)
	return reply, err
}

func ReplyExists(conn *pgxpool.Pool, reply *Reply) bool {
	var exists bool
	statement := "SELECT EXISTS(SELECT 1 FROM replies WHERE tweet_id=$1 AND user_replied_to_id=$2)"
	err := conn.QueryRow(context.Background(), statement, reply.TweetID, reply.ReplyID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
