package models

import (
	"context"

	pgx "github.com/jackc/pgx/v4"
)

type Mention struct {
	ID      int64 `json:"id"`
	TweetID int64 `json:"tweet_id"`
	UserID  int64 `json:"user_id"`
}

//InsertMention inserts a Mention object into the database.  No checking.
func InsertMention(conn *pgx.Conn, mention *Mention) error {
	statement := "INSERT INTO mentions(tweet_id, user_id) VALUES($1, $2)"
	_, err := conn.Exec(context.Background(), statement, mention.TweetID, mention.UserID)
	return err
}

//GetMentionByTID returns a Mention object from the database if they exist.  Otherwise, it returns nil.
//TODO: Return multiple mentions
func GetMentionByTID(conn *pgx.Conn, ID int64) (Mention, error) {
	var mention Mention
	var err error
	statement := "SELECT * FROM mentions WHERE tweet_id=$1"
	err = conn.QueryRow(context.Background(), statement, ID).Scan(&mention.TweetID, &mention.UserID)
	return mention, err
}

//MentionExists checks if a mention exists in the database.
func MentionExists(conn *pgx.Conn, mention *Mention) bool {
	var exists bool
	statement := "SELECT EXISTS(SELECT 1 FROM mentions WHERE tweet_id=$1 AND user_id=$2)"
	err := conn.QueryRow(context.Background(), statement, mention.TweetID, mention.UserID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
