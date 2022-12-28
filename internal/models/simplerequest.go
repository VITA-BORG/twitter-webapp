package models

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

// SimpleRequest is a simple request object.
type SimpleRequest struct {
	ID                 int64  `json:"id"`
	UID                int64  `json:"user_id"`
	Username           string `json:"username"`
	Scrape_connections bool   `json:"scrape_connections"`
}

// InsertSimpleRequest inserts a SimpleRequest object into the database.  No checking. Returns the ID of the inserted row.
func InsertSimpleRequest(conn *pgxpool.Pool, request *SimpleRequest, table string) (int64, error) {

	statement := "INSERT INTO " + table + "(user_id, username, scrape_connections) VALUES($1, $2, $3)"
	tag, err := conn.Exec(context.Background(), statement, request.UID, request.Username, request.Scrape_connections)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// GetSimpleRequests gets all SimpleRequest objects from the database.
func GetSimpleRequests(conn *pgxpool.Pool, follow_status string) ([]SimpleRequest, error) {

	var table_name string

	if follow_status == "follows" {
		table_name = "follow_requests"
	} else if follow_status == "followers" {
		table_name = "follower_requests"
	} else {
		return nil, nil
	}

	var requests []SimpleRequest
	statement := "SELECT * FROM " + table_name
	rows, err := conn.Query(context.Background(), statement)
	if err != nil {
		return requests, err
	}
	defer rows.Close()
	for rows.Next() {
		var request SimpleRequest
		err = rows.Scan(&request.ID, &request.UID, &request.Username, &request.Scrape_connections)
		if err != nil {
			return requests, err
		}
		requests = append(requests, request)
	}
	return requests, nil
}

// DeleteSimpleRequest deletes a SimpleRequest object from the database.
func DeleteSimpleRequest(conn *pgxpool.Pool, request *SimpleRequest, table string) error {
	statement := "DELETE FROM " + table + " WHERE id = $1"
	_, err := conn.Exec(context.Background(), statement, request.ID)
	return err
}
