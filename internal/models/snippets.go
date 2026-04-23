package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type SnippetModel struct {
	DB *sqlx.DB
}

type SnippetModelInterface interface {
	Insert(title string, content string, expires int) (int, error)
	Get(id int) (*Snippet, error)
	Latest() ([]*Snippet, error)
}

func (m *SnippetModel) Insert(title string, content string, expires int) (int, error) {
	res, err := m.DB.Exec(`INSERT INTO snippets (title, content, created, expires)VALUES(?, ? , UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`, title, content, expires)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (m *SnippetModel) Get(id int) (*Snippet, error) {
	s := &Snippet{}
	err := m.DB.Get(s, `SELECT id, title, content, created, expires FROM snippets WHERE expires > UTC_TIMESTAMP() AND id = ?`, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}
func (m *SnippetModel) Latest() ([]*Snippet, error) {
	snippets := []*Snippet{}
	err := m.DB.Select(&snippets, `SELECT id, title, content, created, expires FROM snippets WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`)
	if err != nil {
		return nil, err
	}

	return snippets, nil
}
