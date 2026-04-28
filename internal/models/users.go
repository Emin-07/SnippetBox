package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *sqlx.DB
}

type UserModelInterface interface {
	Insert(name string, email string, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	Get(id int) (*User, error)
	ChangePassword(id int, currentPassword, newPassword string) error
}

func (m *UserModel) Insert(name string, email string, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	_, err = m.DB.Exec(`INSERT INTO users (name, email, hashed_password, created)VALUES(?, ? , ?, UTC_TIMESTAMP())`, name, email, hashedPassword)
	if err != nil {
		var mySQLError *mysql.MySQLError
		if errors.As(err, &mySQLError) {
			if mySQLError.Number == 1062 && strings.Contains(mySQLError.Message, "users_uc_email") {
				return ErrDuplicateEmail
			}
		}
		return err
	}
	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {

	var id int
	var hashedPassword []byte
	err := m.DB.QueryRow("SELECT id, hashed_password FROM users WHERE email = ?", email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	return id, nil
}
func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool
	stmt := "SELECT EXISTS(SELECT true FROM users WHERE id = ?)"
	err := m.DB.QueryRow(stmt, id).Scan(&exists)
	return exists, err
}

func (m *UserModel) Get(id int) (*User, error) {
	user := User{}
	err := m.DB.Get(&user, "SELECT id, name, email, created  FROM users WHERE id = ?", id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return &user, nil
}

func (m *UserModel) ChangePassword(id int, currentPassword, newPassword string) error {
	var userHashPassword string
	err := m.DB.QueryRow("SELECT hashed_password FROM users WHERE id = ?", id).Scan(&userHashPassword)
	if err != nil {
		return err
	}
	if err = bcrypt.CompareHashAndPassword([]byte(userHashPassword), []byte(currentPassword)); err != nil {
		return err
	}
	newHashedPassword, err := (bcrypt.GenerateFromPassword([]byte(newPassword), 12))
	if err != nil {
		return err
	}
	_, err = m.DB.Exec("UPDATE users SET hashed_password = ? WHERE id = ?", newHashedPassword, id)
	if err != nil {
		return err
	}

	return nil
}
