package models

import (
	"context"
	"time"

	pgx "github.com/jackc/pgx/v4"
)

type User struct {
	Id          int        `json:"id"`
	ProfileName string     `json:"profile_name"`
	Handle      string     `json:"handle"`
	Gender      string     `json:"gender"`
	IsPerson    bool       `json:"is_person"`
	Joined      *time.Time `json:"joined"`
	Bio         string     `json:"bio"`
	Location    string     `json:"location"`
	Verified    bool       `json:"verified"`
	Avatar      string     `json:"avatar"`
	Tweets      int        `json:"tweets"`
	Likes       int        `json:"likes"`
	Media       int        `json:"media"`
	Following   int        `json:"following"`
	Followers   int        `json:"followers"`
	CollectedAt *time.Time `json:"collected_at"`
}

var format string = "2006-01-02"

//insert_user inserts a User object into the database.  No checking.
func insert_user(conn *pgx.Conn, user User) error {
	statement := "INSERT INTO users(id, profile_name, handle, gender, is_person, joined, bio, location, verified, avatar, tweets, likes, media, following, followers, collected_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)"
	_, err := conn.Exec(context.Background(), statement, user.Id, user.ProfileName, user.Handle, user.Gender, user.IsPerson, user.Joined.Format(format), user.Bio, user.Location, user.Verified, user.Avatar, user.Tweets, user.Likes, user.Media, user.Following, user.Followers, user.CollectedAt.Format(format))
	return err
}

//get_user gets a User object from the database if they exist.  Otherwise, it returns nil.
func get_user(conn *pgx.Conn, handle string) (User, error) {
	var user User
	var err error
	statement := "SELECT * FROM users WHERE handle=$1"
	err = conn.QueryRow(context.Background(), statement, handle).Scan(&user.Id, &user.ProfileName, &user.Handle, &user.Gender, &user.IsPerson, &user.Joined, &user.Bio, &user.Location, &user.Verified, &user.Avatar, &user.Likes, &user.Media, &user.Following, &user.Followers, &user.CollectedAt)
	return user, err
}

//user_exists checks if a user exists in the database.
func user_exists(conn *pgx.Conn, handle string) bool {
	var exists bool
	statement := "SELECT EXISTS(SELECT 1 FROM users WHERE handle=$1)"
	err := conn.QueryRow(context.Background(), statement, handle).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
