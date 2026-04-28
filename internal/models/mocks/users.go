package mocks

import (
	"time"

	"snippetbox.net/internal/models"
)

type UserModel struct{}

func (m *UserModel) Insert(name string, email string, password string) error {
	switch email {
	case "dupe@email.com", "mohmad@email.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	if email == "mocker@email.com" && password == "password" {
		return 1, nil
	}
	return 0, models.ErrInvalidCredentials

}
func (m *UserModel) Exists(id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (m *UserModel) Get(id int) (*models.User, error) {
	if id == 1 {
		return &models.User{ID: 1,
			Name:           "Mocker",
			Email:          "mocker@email.com",
			HashedPassword: []byte("$$2a$12$rvsOpIiQwzmp/r2OF4sHjemIQoXOx3YrtPRF0zvCCajaU0AxqIYPu"),
			Created:        time.Date(2022, time.January, 1, 10, 0, 0, 0, time.UTC)}, nil
	} else {
		return nil, models.ErrNoRecord
	}
}

func (m *UserModel) ChangePassword(id int, currentPassword, newPassword string) error {
	if id != 1 {
		return models.ErrNoRecord
	}
	return nil
}
