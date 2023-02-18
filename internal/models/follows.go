package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Follow struct {
	ID               int64     `json:"id"`
	FollowerID       int64     `json:"follower_id"`
	FolloweeID       int64     `json:"followee_id"`
	FollowerUsername string    `json:"follower_username"`
	FolloweeUsername string    `json:"followee_username"`
	CreatedAt        time.Time `json:"created_at"`
	CollectedAt      time.Time `json:"collected_at"`
}

// InsertFollow inserts a Follow object into the database.
func InsertFollow(conn *pgxpool.Pool, follow *Follow) error {
	if FollowExists(conn, follow) {
		return nil
	}
	statement := "INSERT INTO follows(follower_id, followee_id, created_at, collected_at) VALUES($1, $2, $3, $4)"
	_, err := conn.Exec(context.Background(), statement, follow.FollowerID, follow.FolloweeID, follow.CreatedAt, follow.CollectedAt)
	return err
}

// GetFollowers retrieves all followers of a user returns a slice of pointers to Follow objects from the database if they exist.  Otherwise, it returns nil.
func GetFollowers(conn *pgxpool.Pool, uid int64) ([]*Follow, error) {
	var follows []*Follow
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

		follow.FollowerUsername, err = GetUsernameByID(conn, follow.FollowerID)
		if err != nil {
			return nil, err
		}

		follow.FolloweeUsername, err = GetUsernameByID(conn, follow.FolloweeID)
		if err != nil {
			return nil, err
		}

		follows = append(follows, &follow)
	}
	return follows, nil
}

// GetFollows retrieves all follows of a user and returns a slice of pointers to Follow objects from the database if they exist.  Otherwise, it returns nil.
func GetFollows(conn *pgxpool.Pool, uid int64) ([]*Follow, error) {
	var follows []*Follow
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

		follow.FollowerUsername, err = GetUsernameByID(conn, follow.FollowerID)
		if err != nil {
			return nil, err
		}

		follow.FolloweeUsername, err = GetUsernameByID(conn, follow.FolloweeID)
		if err != nil {
			return nil, err
		}

		follows = append(follows, &follow)
	}
	return follows, nil
}

// FollowExists checks if a follow exists in the database.  Returns true if it does.  Otherwise, it returns false.
func FollowExists(conn *pgxpool.Pool, follow *Follow) bool {
	var exists bool
	statement := "SELECT EXISTS(SELECT 1 FROM follows WHERE follower_id=$1 AND followee_id=$2)"
	err := conn.QueryRow(context.Background(), statement, follow.FollowerID, follow.FolloweeID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

// AddFollows takes a slice of pointers to Follow objects and adds them to the database if they do not already exist.
func AddFollows(conn *pgxpool.Pool, follows []*Follow) error {
	for _, follow := range follows {
		if !FollowExists(conn, follow) {
			err := InsertFollow(conn, follow)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
