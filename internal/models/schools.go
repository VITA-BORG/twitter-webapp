package models

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type School struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	TopRated bool   `json:"top_rated"`
	Public   bool   `json:"public"`
	City     string `json:"city"`
	State    string `json:"state_province"`
	Country  string `json:"country"`
	User_ID  int64  `json:"user_id"`
}

//InsertSchool inserts a School object into the database.  No checking.
func InsertSchool(conn *pgxpool.Pool, school *School) error {
	statement := "INSERT INTO schools(id, name, top_rated, public, city, state_province, country, user_id) VALUES($1, $2, $3, $4, $5, $6, $7, $8)"
	_, err := conn.Exec(context.Background(), statement, school.ID, school.Name, school.TopRated, school.Public, school.City, school.State, school.Country, school.User_ID)
	return err
}

//GetSchoolByID returns a School object from the database if they exist.  Otherwise, it returns nil.
func GetSchoolByID(conn *pgxpool.Pool, ID int) (*School, error) {
	var school School
	var err error
	statement := "SELECT * FROM schools WHERE id=$1"
	err = conn.QueryRow(context.Background(), statement, ID).Scan(&school.ID, &school.Name, &school.TopRated, &school.Public, &school.City, &school.State, &school.Country, &school.User_ID)
	return &school, err
}

//GetSchoolByName returns a School object from the database if they exist.  Otherwise, it returns nil.
func GetSchoolByName(conn *pgxpool.Pool, name string) (*School, error) {
	var school School
	var err error
	statement := "SELECT * FROM schools WHERE name=$1"
	err = conn.QueryRow(context.Background(), statement, name).Scan(&school.ID, &school.Name, &school.TopRated, &school.Public, &school.City, &school.State, &school.Country, &school.User_ID)
	return &school, err
}

//GetSchoolIDByName returns the ID of a school from the database if they exist.  Otherwise, it returns nil.
func GetSchoolIDByName(conn *pgxpool.Pool, name string) (int, error) {
	var ID int
	var err error
	statement := "SELECT id FROM schools WHERE name=$1"
	err = conn.QueryRow(context.Background(), statement, name).Scan(&ID)
	return ID, err
}

//SchoolExists checks if a school exists in the database.
func SchoolExists(conn *pgxpool.Pool, ID int64) bool {
	var exists bool
	statement := "SELECT EXISTS(SELECT 1 FROM schools WHERE id=$1)"
	err := conn.QueryRow(context.Background(), statement, ID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

//SchoolUserIDExists checks if a school user ID exists in the database.
func SchoolUserIDExists(conn *pgxpool.Pool, ID int64) bool {
	var exists bool
	statement := "SELECT EXISTS(SELECT 1 FROM schools WHERE user_id=$1)"
	err := conn.QueryRow(context.Background(), statement, ID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

//NumberOfSchools returns the number of schools in the database.
func NumberOfSchools(conn *pgxpool.Pool) (int, error) {
	var count int
	statement := "SELECT COUNT(*) FROM schools"
	err := conn.QueryRow(context.Background(), statement).Scan(&count)
	return count, err
}

//GetAllSchools returns a slice of all schools in the database.
func GetAllSchools(conn *pgxpool.Pool) ([]School, error) {
	var schools []School
	var err error
	statement := "SELECT * FROM schools"
	rows, err := conn.Query(context.Background(), statement)
	if err != nil {
		return schools, err
	}
	for rows.Next() {
		var school School
		err = rows.Scan(&school.ID, &school.Name, &school.TopRated, &school.Public, &school.City, &school.State, &school.Country, &school.User_ID)
		if err != nil {
			return schools, err
		}
		schools = append(schools, school)
	}
	return schools, err
}
