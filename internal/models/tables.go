package models

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

var tables = []string{"users", "tweets", "schools", "students", "replies", "mentions", "bio_tags", "hashtags", "follows", "sessions", "admins", "follow_requests", "follower_requests"}

// DeleteTables drops all tables in the database.  Only use when testing or when you want to start from scratch.
func DeleteTables(conn *pgxpool.Pool) error {
	var statement string
	for _, table := range tables {
		statement = fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
		_, err := conn.Exec(context.Background(), statement)
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

// CreateTables creates all tables in the database.  Only use when testing or when you want to start from scratch.
func CreateTables(conn *pgxpool.Pool) error {

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
        collected_at timestamp,
		is_participant boolean
		)`
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
		)`
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
		)`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	//Creates student table
	statement = `create table students(
		school_id int references schools(id) ON DELETE CASCADE,
		user_id bigint references users(id) ON DELETE CASCADE,
		cohort int
	)`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	//Creates replies table
	statement = `create table replies(
		id serial primary key,
        tweet_id bigint references tweets(id) ON DELETE CASCADE,
        user_replied_to_id bigint references users(id)
		)`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	//Creates mentions table
	statement = `create table mentions(
		id serial primary key,
        tweet_id bigint references tweets(id) ON DELETE CASCADE,
        user_id bigint references users(id)
		)`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	//Creates bio_tags table
	statement = `create table bio_tags(
		id serial primary key,
        user_id bigint references users(id),
        mentioned_user_id bigint references users(id),
        collected_at timestamp
		)`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	//Creates hashtags table
	statement = `create table hashtags(
		id serial primary key,
        tag varchar(512),
        tweet_id bigint references tweets(id) ON DELETE CASCADE
		)`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	//Creates follows table
	statement = `create table follows(
		id serial primary key,
        follower_id bigint references users(id),
        followee_id bigint references users(id),
        created_at timestamp,
        collected_at timestamp
		)`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	//Creates sessions table
	statement = `create table sessions(
		token text primary key,
		data bytea NOT NULL,
		expiry timestamptz NOT NULL
		)`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	statement = `create table follow_requests(
		id serial primary key,
		user_id bigint,
		username varchar(256),
		scrape_connections boolean
		)`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	statement = `create table follower_requests(
		id serial primary key,
		user_id bigint,
		username varchar(256),
		scrape_connections boolean
		)`
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	statement = "CREATE INDEX sessions_expiry ON sessions (expiry)"
	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	statement = `CREATE TABLE admins (
		id serial primary key,
		name varchar(256),
		email varchar(256) unique,
		password varchar(256),
		created_at timestamp
		)`

	_, err = conn.Exec(context.Background(), statement)
	if err != nil {
		return err
	}

	return nil
}
