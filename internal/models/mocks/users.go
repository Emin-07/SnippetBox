package mocks

import (
	"snippetbox.net/internal/models"
)

type UserModel struct{}

func (m *UserModel) Insert(name string, email string, password string) error {
	switch email {
	case "mocker@email.com", "mohmad@email.com":
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
