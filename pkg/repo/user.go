package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
)

/*----------------------------------- User Type ----------------------------------- */

type UserCore struct {
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

type User struct {
	UserCore
	Role          string `json:"role"`
	FullName      string `json:"full_name,omitempty"`
	Username      string `json:"username,omitempty"`
	DateOfBirth   string `json:"date_of_birth"`
	Gender        string `json:"gender,omitempty"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	AccountStatus string `json:"account_status"`
	ImageUrl      string `json:"image_url"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	Id            string `json:"id"`
}

func (repo *Repo) GetUserById(ctx context.Context, userId string) (*User, error) {
	user := new(User)
	err := repo.db.QueryRowContext(ctx, `SELECT id, role, email, password_hash, COALESCE(username, ''), COALESCE(full_name, ''), COALESCE(date_of_birth, '-infinity'), COALESCE(gender, ''), COALESCE(phone_number, ''), COALESCE(account_status, ''), COALESCE(image_url, ''), created_at, updated_at FROM users WHERE id=$1 LIMIT 1;`, userId).Scan(&user.Id, &user.Role, &user.Email, &user.PasswordHash, &user.Username, &user.FullName, &user.DateOfBirth, &user.Gender, &user.PhoneNumber, &user.AccountStatus, &user.ImageUrl, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		if err, ok := err.(*pq.Error); ok {
			switch code := err.Code.Name(); code {
			case "undefined_column":
				return nil, ErrUserNotFound
			default:
				return nil, errors.New(code)
			}
		}
		return nil, err
	}
	return user, nil
}

func (repo *Repo) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	user := new(User)
	err := repo.db.QueryRowContext(ctx, `SELECT id, role, email, password_hash, COALESCE(username, ''), COALESCE(full_name, ''), COALESCE(date_of_birth, '-infinity'), COALESCE(gender, ''), COALESCE(phone_number, ''), COALESCE(account_status, ''), COALESCE(image_url, ''), created_at, updated_at FROM users WHERE email=$1 LIMIT 1;`, email).Scan(&user.Id, &user.Role, &user.Email, &user.PasswordHash, &user.Username, &user.FullName, &user.DateOfBirth, &user.Gender, &user.PhoneNumber, &user.AccountStatus, &user.ImageUrl, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		if err, ok := err.(*pq.Error); ok {
			switch code := err.Code.Name(); code {
			case "undefined_column":
				return nil, ErrUserNotFound
			default:
				return nil, errors.New(code)
			}
		}
		return nil, err
	}
	return user, nil
}

func (repo *Repo) CreateUser(ctx context.Context, user *UserCore) (string, error) {
	var id string
	err := repo.db.QueryRowContext(ctx, `INSERT INTO users(email, password_hash) VALUES($1, $2) RETURNING id;`, user.Email, user.PasswordHash).Scan(&id)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			switch code := err.Code.Name(); code {
			case "unique_violation":
				return "", ErrUserAlreadyExists
			default:
				return "", errors.New(code)
			}
		}
		return "", err
	}
	return id, err
}

func (repo *Repo) DeleteUserById(ctx context.Context, id string) error {
	_, err := repo.db.ExecContext(ctx, `DELETE FROM users WHERE id=$1;`, id)
	return err
}

func (repo *Repo) Update(ctx context.Context, id string, updates map[string]any) error {
	query := "UPDATE users SET "
	var params []interface{}

	count := 1
	for key, value := range updates {
		query += fmt.Sprintf("%s=$%v, ", key, count)
		params = append(params, value)
		count++
	}

	// Remove the trailing comma and space
	query = query[:len(query)-2]

	query += fmt.Sprintf(" WHERE id=$%v;", count)
	params = append(params, id)
	_, err := repo.db.Exec(query, params...)
	return err
}
