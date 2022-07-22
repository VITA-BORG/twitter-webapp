package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type BioTag struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	MentionedUserID int64      `json:"mentioned_user_id"`
	CollectedAt     *time.Time `json:"collected_at"`
}

//InsertBioTag inserts a BioTag object into the database.  No checking.
func InsertBioTag(conn *pgxpool.Pool, bioTag *BioTag) error {
	statement := "INSERT INTO bio_tags(user_id, mentioned_user_id, collected_at) VALUES($1, $2, $3)"
	_, err := conn.Exec(context.Background(), statement, bioTag.UserID, bioTag.MentionedUserID, bioTag.CollectedAt)
	return err
}

//TagExists checks if a tag exists in the database given a UserID and MentionedUserID
func TagExists(conn *pgxpool.Pool, userID int64, mentionedUserID int64) bool {
	var exists bool
	statement := "SELECT EXISTS(SELECT 1 FROM bio_tags WHERE user_id=$1 AND mentioned_user_id=$2)"
	err := conn.QueryRow(context.Background(), statement, userID, mentionedUserID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
