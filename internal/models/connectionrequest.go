package models

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

// ConnectionRequest is a connection request object.
type ConnectionRequest struct {
	ID                 int64  `json:"id"`
	UID                int64  `json:"user_id"`
	Username           string `json:"username"`
	FollowsOrFollowers string `json:"follows_or_followers"`
}

// InsertConnectionRequest inserts a ConnectionRequest object into the database.  No checking.
func InsertConnectionRequest(conn *pgxpool.Pool, request *ConnectionRequest) (int64, error) {

	statement := "INSERT INTO connection_requests(user_id, username, follows_or_followers) VALUES($1, $2, $3)"
	tag, err := conn.Exec(context.Background(), statement, request.UID, request.Username, request.FollowsOrFollowers)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// GetConnectionRequests gets all ConnectionRequest objects from the database.
func GetConnectionRequests(conn *pgxpool.Pool) ([]*ConnectionRequest, error) {

	var requests []*ConnectionRequest
	statement := "SELECT * FROM connection_requests"
	rows, err := conn.Query(context.Background(), statement)
	if err != nil {
		return requests, err
	}
	defer rows.Close()
	for rows.Next() {
		var request ConnectionRequest
		err = rows.Scan(&request.ID, &request.UID, &request.Username, &request.FollowsOrFollowers)
		if err != nil {
			return requests, err
		}
		requests = append(requests, &request)
	}
	return requests, nil
}

// DeleteConnectionRequest deletes a ConnectionRequest object from the database.
func DeleteConnectionRequest(conn *pgxpool.Pool, requestID int64) error {
	statement := "DELETE FROM connection_requests WHERE id = $1"
	_, err := conn.Exec(context.Background(), statement, requestID)
	return err
}
