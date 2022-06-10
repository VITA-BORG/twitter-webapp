package models

import (
	"context"

	pgx "github.com/jackc/pgx/v4"
)

type School struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	TopRated bool   `json:"top_rated"`
	Public   bool   `json:"public"`
	City     string `json:"city"`
	State    string `json:"state_province"`
	Country  string `json:"country"`
	User_ID  int64  `json:"user_id"`
}

//InsertSchool inserts a School object into the database.  No checking.
func InsertSchool(conn *pgx.Conn, school School) error {
	statement := "INSERT INTO schools(id, name, top_rated, public, city, state_province, country, user_id) VALUES($1, $2, $3, $4, $5, $6, $7, $8)"
	_, err := conn.Exec(context.Background(), statement, school.ID, school.Name, school.TopRated, school.Public, school.City, school.State, school.Country, school.User_ID)
	return err
}

//GetSchoolByID returns a School object from the database if they exist.  Otherwise, it returns nil.
func GetSchoolByID(conn *pgx.Conn, ID int64) (*School, error) {
	var school School
	var err error
	statement := "SELECT * FROM schools WHERE id=$1"
	err = conn.QueryRow(context.Background(), statement, ID).Scan(&school.ID, &school.Name, &school.TopRated, &school.Public, &school.City, &school.State, &school.Country, &school.User_ID)
	return &school, err
}

//GetSchoolByName returns a School object from the database if they exist.  Otherwise, it returns nil.
func GetSchoolByName(conn *pgx.Conn, name string) (*School, error) {
	var school School
	var err error
	statement := "SELECT * FROM schools WHERE name=$1"
	err = conn.QueryRow(context.Background(), statement, name).Scan(&school.ID, &school.Name, &school.TopRated, &school.Public, &school.City, &school.State, &school.Country, &school.User_ID)
	return &school, err
}

//SchoolExists checks if a school exists in the database.
func SchoolExists(conn *pgx.Conn, ID int64) bool {
	var exists bool
	statement := "SELECT EXISTS(SELECT 1 FROM schools WHERE id=$1)"
	err := conn.QueryRow(context.Background(), statement, ID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
