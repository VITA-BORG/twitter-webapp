package models

import (
	"context"

	pgx "github.com/jackc/pgx/v4"
)

var tables = []string{"users", "tweets", "schools", "students", "replies", "mentions", "bio_tags", "hashtags", "follows"}

//ResetTables resets all tables in the database.  Only use when testing or when you want to start from scratch.
func ResetTables(conn *pgx.Conn) error {
	statement := "DROP TABLE IF EXISTS $1 CASCADE"
	for _, table := range tables {
		_, err := conn.Exec(context.Background(), statement, table)
		if err != nil {
			return err
		}
	}

	statement = "DROP TYPE IF EXISTS gender CASCADE"
	_, err := conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}
	return nil
}

//CreateTables creates all tables in the database.  Only use when testing or when you want to start from scratch.
func CreateTables(conn *pgx.Conn) error {

	statement := "CREATE TYPE gender AS ENUM ('M', 'F', 'X')"
	_, err := conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	//Creates user table
	statement = `create table users(
		id bigint primary key,
		profile_name varchar(256),
		handle varchar(64),
		gender gender,
        is_person boolean,
        joined timestamp,
        bio text,
        location varchar(256),
        verified boolean,
        avatar varchar(512),
        tweets int,
        likes int,
        media int,
        following int,
        followers int,
        collected_at timestamp`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	//Creates tweet table
	statement = `create table tweets(
		id bigint primary key,
        conversation_id bigint references tweets(id) ON DELETE CASCADE,
        text text,
        posted_at timestamp,
        url varchar (256),
        user_id bigint references users(id),
        is_retweet boolean,
        retweet_id bigint references tweets(id) ON DELETE CASCADE,
        likes int,
        retweets int,
        replies int,
        collected_at timestamp
		`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	//Creates school table
	statement = `create table schools(
		id int primary key,
        name varchar(256),
        top_rated boolean,
        public boolean,
        city varchar(128),
        state_province varchar(4),
        country varchar(4),
        user_id bigint references users(id)
		`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}
	return nil
}
