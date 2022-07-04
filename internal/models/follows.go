package models

import (
	"context"
	"time"

	pgx "github.com/jackc/pgx/v4"
)

type Follow struct {
	ID          int64     `json:"id"`
	FollowerID  int64     `json:"follower_id"`
	FolloweeID  int64     `json:"followee_id"`
	CreatedAt   time.Time `json:"created_at"`
	CollectedAt time.Time `json:"collected_at"`
}

//InsertFollow inserts a Follow object into the database.  No checking.
func InsertFollow(conn *pgx.Conn, follow Follow) error {
	statement := "INSERT INTO follows(follower_id, followee_id, created_at, collected_at) VALUES($1, $2, $3, $4)"
	_, err := conn.Exec(context.Background(), statement, follow.FollowerID, follow.FolloweeID, follow.CreatedAt, follow.CollectedAt)
	return err
}

//GetFollowers retrieves all followers of a user returns a slice of Follow objects from the database if they exist.  Otherwise, it returns nil.
func GetFollowers(conn *pgx.Conn, uid int64) ([]Follow, error) {
	var follows []Follow
	var err error
	statement := "SELECT * FROM follows WHERE followee_id=$1"
	rows, err := conn.Query(context.Background(), statement, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var follow Follow
		err = rows.Scan(&follow.ID, &follow.FollowerID, &follow.FolloweeID, &follow.CreatedAt, &follow.CollectedAt)
		if err != nil {
			return nil, err
		}
		follows = append(follows, follow)
	}
	return follows, nil
}

//GetFollows retrieves all follows of a user and returns a slice of Follow objects from the database if they exist.  Otherwise, it returns nil.
func GetFollows(conn *pgx.Conn, uid int64) ([]Follow, error) {
	var follows []Follow
	var err error
	statement := "SELECT * FROM follows WHERE follower_id=$1"
	rows, err := conn.Query(context.Background(), statement, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var follow Follow
		err = rows.Scan(&follow.ID, &follow.FollowerID, &follow.FolloweeID, &follow.CreatedAt, &follow.CollectedAt)
		if err != nil {
			return nil, err
		}
		follows = append(follows, follow)
	}
	return follows, nil
}
