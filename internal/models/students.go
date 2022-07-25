package models

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Student struct {
	SchoolID int   `json:"school_id"`
	Cohort   int   `json:"cohort"`
	UserID   int64 `json:"user_id"`
}

//InsertStudent inserts a Student object into the database.  No checking.
func InsertStudent(conn *pgxpool.Pool, student *Student) error {
	statement := "INSERT INTO students(school_id, cohort, user_id) VALUES($1, $2, $3)"
	_, err := conn.Exec(context.Background(), statement, student.SchoolID, student.Cohort, student.UserID)
	return err
}

//GetStudentByID returns a Student object from the database if they exist.  Otherwise, it returns nil.
func GetStudentByID(conn *pgxpool.Pool, ID int64) (*Student, error) {
	var student Student
	var err error
	statement := "SELECT * FROM students WHERE user_id=$1"
	err = conn.QueryRow(context.Background(), statement, ID).Scan(&student.SchoolID, &student.Cohort, &student.UserID)
	return &student, err
}

//StudentExists checks if a student exists in the database.
func StudentExists(conn *pgxpool.Pool, ID int64) bool {
	var exists bool
	statement := "SELECT EXISTS(SELECT 1 FROM students WHERE user_id=$1)"
	err := conn.QueryRow(context.Background(), statement, ID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}
