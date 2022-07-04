package models

import (
	"context"
	"time"

	pgx "github.com/jackc/pgx/v4"
)

type BioTag struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	MentionedUserID int64      `json:"mentioned_user_id"`
	CollectedAt     *time.Time `json:"collected_at"`
}

//InsertBioTag inserts a BioTag object into the database.  No checking.
func InsertBioTag(conn *pgx.Conn, bioTag BioTag) error {
	statement := "INSERT INTO biotags(user_id, mentioned_user_id, collected_at) VALUES($1, $2, $3)"
	_, err := conn.Exec(context.Background(), statement, bioTag.UserID, bioTag.MentionedUserID, bioTag.CollectedAt)
	return err
}

//TODO GetBioTagByID, returns multiple mentions
