package models

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	//Hashed Password
	Password  []byte     `json:"password"`
	CreatedAt *time.Time `json:"created_at"`
}

//InsertAdmin inserts a Admin object into the database.  No checking.
func InsertAdmin(conn *pgxpool.Pool, admin *Admin) error {

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	statement := "INSERT INTO admins(name, email, password, created_at) VALUES($1, $2, $3, $4)"
	_, err = conn.Exec(context.Background(), statement, admin.Name, admin.Email, hashedPassword, admin.CreatedAt)
	if err != nil {
		//check if email already exists
		if strings.Contains(err.Error(), "duplicate key value") {
			return errors.New("email already exists")
		}
		return err
	}
	return nil
}

//AuthenticateAdmin checks if an admin password pair exists and is valid
//Returns the ID if it is valid
func AuthenticateAdmin(conn *pgxpool.Pool, email string, password string) (int, error) {
	return 0, nil
}

//AdminExists checks if an admin exists in the database.
func AdminExists(conn *pgxpool.Pool, ID int) bool {
	return false
}
