package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type User struct {
	ID            int64      `json:"id"`
	ProfileName   string     `json:"profile_name"`
	Handle        string     `json:"handle"`
	Gender        *string    `json:"gender"`
	IsPerson      bool       `json:"is_person"`
	Joined        *time.Time `json:"joined"`
	Bio           string     `json:"bio"`
	Location      string     `json:"location"`
	Verified      bool       `json:"verified"`
	Avatar        string     `json:"avatar"`
	Tweets        int        `json:"tweets"`
	Likes         int        `json:"likes"`
	Media         int        `json:"media"`
	Following     int        `json:"following"`
	Followers     int        `json:"followers"`
	CollectedAt   *time.Time `json:"collected_at"`
	IsParticipant bool       `json:"is_participant"`
}

var Format string = "2006-01-02"

// InsertUser inserts a User object into the database.
func InsertUser(conn *pgxpool.Pool, user *User) error {
	if UserIDExists(conn, user.ID) {
		return nil
	}
	statement := "INSERT INTO users(id, profile_name, handle, gender, is_person, joined, bio, location, verified, avatar, tweets, likes, media, following, followers, collected_at, is_participant) VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)"
	_, err := conn.Exec(context.Background(), statement, user.ID, user.ProfileName, user.Handle, user.Gender, user.IsPerson, user.Joined, user.Bio, user.Location, user.Verified, user.Avatar, user.Tweets, user.Likes, user.Media, user.Following, user.Followers, user.CollectedAt, user.IsParticipant)
	return err
}

// GetUserByHandle returns a User object from the database if they exist.  Otherwise, it returns nil.
func GetUserByHandle(conn *pgxpool.Pool, handle string) (*User, error) {
	var user User
	var err error
	statement := "SELECT * FROM users WHERE handle ILIKE $1"
	err = conn.QueryRow(context.Background(), statement, handle).Scan(&user.ID, &user.ProfileName, &user.Handle, &user.Gender, &user.IsPerson, &user.Joined, &user.Bio, &user.Location, &user.Verified, &user.Avatar, &user.Tweets, &user.Likes, &user.Media, &user.Following, &user.Followers, &user.CollectedAt, &user.IsParticipant)
	return &user, err
}

// GetUserByID returns a User object from the database if they exist.  Otherwise, it returns nil.
func GetUserByID(conn *pgxpool.Pool, ID int64) (*User, error) {
	var user User
	var err error
	statement := "SELECT * FROM users WHERE id=$1"
	err = conn.QueryRow(context.Background(), statement, ID).Scan(&user.ID, &user.ProfileName, &user.Handle, &user.Gender, &user.IsPerson, &user.Joined, &user.Bio, &user.Location, &user.Verified, &user.Avatar, &user.Tweets, &user.Likes, &user.Media, &user.Following, &user.Followers, &user.CollectedAt, &user.IsParticipant)
	return &user, err
}

// UserExists checks if a user exists in the database.
func UserExists(conn *pgxpool.Pool, handle string) bool {
	var exists bool
	statement := "SELECT EXISTS(SELECT 1 FROM users WHERE handle ILIKE $1)"
	err := conn.QueryRow(context.Background(), statement, handle).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

// UserIDExists checks if a user ID exists in the database.
func UserIDExists(conn *pgxpool.Pool, ID int64) bool {
	var exists bool
	statement := "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)"
	err := conn.QueryRow(context.Background(), statement, ID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func GetUsernameByID(conn *pgxpool.Pool, ID int64) (string, error) {
	var username string
	var err error
	statement := "SELECT handle FROM users WHERE id=$1"
	err = conn.QueryRow(context.Background(), statement, ID).Scan(&username)
	return username, err
}

// GetUserIDByHandle returns the user's ID given their handle
func GetUserIDByHandle(conn *pgxpool.Pool, handle string) (int64, error) {
	var id int64
	var err error
	statement := "SELECT id FROM users where handle ILIKE $1"
	err = conn.QueryRow(context.Background(), statement, handle).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpdateUser updates the user's record given a user struct
func UpdateUser(conn *pgxpool.Pool, user *User) error {
	statement := "UPDATE users SET profile_name=$1, handle=$2, gender=$3, is_person=$4, joined=$5, bio=$6, location=$7, verified=$8, avatar=$9, tweets=$10, likes=$11, media=$12, following=$13, followers=$14, collected_at=$15, is_participant=$16 WHERE id=$17"
	_, err := conn.Exec(context.Background(), statement, user.ProfileName, user.Handle, user.Gender, user.IsPerson, user.Joined, user.Bio, user.Location, user.Verified, user.Avatar, user.Tweets, user.Likes, user.Media, user.Following, user.Followers, user.CollectedAt, user.IsParticipant, user.ID)
	return err

}

// UpdateUserHandle updates the user's handle given a user struct
func UpdateUserHandle(conn *pgxpool.Pool, user *User) error {
	statement := "UPDATE users SET handle=$1 WHERE id=$2"
	_, err := conn.Exec(context.Background(), statement, user.Handle, user.ID)
	return err
}

// GetAllUsernames returns a list of all usernames in the database.
func GetAllUsernames(conn *pgxpool.Pool) ([]string, error) {
	var usernames []string
	var err error
	statement := "SELECT handle FROM users"
	rows, err := conn.Query(context.Background(), statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var handle string
		err = rows.Scan(&handle)
		if err != nil {
			return nil, err
		}
		usernames = append(usernames, handle)
	}
	return usernames, nil
}

// GetAllParticipants returns a list of all participants in the database.
func GetAllParticipants(conn *pgxpool.Pool) ([]User, error) {
	var users []User
	var err error
	statement := "SELECT * FROM users WHERE is_participant=TRUE"
	rows, err := conn.Query(context.Background(), statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.ProfileName, &user.Handle, &user.Gender, &user.IsPerson, &user.Joined, &user.Bio, &user.Location, &user.Verified, &user.Avatar, &user.Tweets, &user.Likes, &user.Media, &user.Following, &user.Followers, &user.CollectedAt, &user.IsParticipant)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

// GetUserCount returns the number of users in the database.
func GetUserCount(conn *pgxpool.Pool) (int, error) {
	var count int
	var err error
	statement := "SELECT COUNT(*) FROM users"
	err = conn.QueryRow(context.Background(), statement).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
