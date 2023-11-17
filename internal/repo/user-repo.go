package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"

	"github.com/lib/pq"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrUserNotFound      = errors.New("user not found")
	UserRepo             = NewUserRepo(postgresDb)
)

type UserCore struct {
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

type User struct {
	UserCore
	Id            uint   `json:"id"`
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
}

func NewUserRepo(db *sql.DB) *UserRepository {
	repo := new(UserRepository)
	repo.db = db
	err := repo.migrate()
	if err != nil {
		log.Fatalln(err)
	}
	return repo
}

type UserRepository struct {
	db *sql.DB
}

func (repo *UserRepository) migrate() error {
	repo.db.Exec("CREATE EXTENSION IF NOT EXISTS CITEXT;")
	_, err := repo.db.Exec(`CREATE TABLE IF NOT EXISTS users(
    id SERIAL PRIMARY KEY,
	role TEXT CHECK (role IN ('user', 'staff', 'admin')) DEFAULT 'user',
    email CITEXT NOT NULL UNIQUE CHECK (LENGTH(email)<=64),
    password_hash TEXT NOT NULL CHECK (LENGTH(email)<=72),
	username TEXT UNIQUE CHECK (LENGTH(username)<=32),
    full_name TEXT CHECK (LENGTH(full_name)<=64),
    date_of_birth DATE,
    gender TEXT CHECK (gender IN ('male', 'female', 'other')),
	phone_number TEXT CHECK (LENGTH(phone_number)<=16),
	account_status TEXT CHECK (account_status IN ('active', 'suspended', 'banned')) DEFAULT 'active',
	image_url TEXT,
    created_at TIMESTAMPTZ DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ DEFAULT current_timestamp
);`)
	return err
}

func (repo *UserRepository) GetById(ctx context.Context, userId uint) (*User, error) {
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

func (repo *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
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

func (repo *UserRepository) GetAll(ctx context.Context, page uint) (*Paginated[*User], error) {
	var pageSize uint = 20
	var users []*User

	countCh := make(chan uint)

	go func() {
		defer close(countCh)
		var count uint
		repo.db.QueryRow("SELECT COUNT(*) FROM users;").Scan(&count)
		countCh <- count
	}()

	rows, err := repo.db.QueryContext(ctx, `SELECT id, role, email, password_hash, COALESCE(username, ''), COALESCE(full_name, ''), COALESCE(date_of_birth, '-infinity'), COALESCE(gender, ''), COALESCE(phone_number, ''), COALESCE(account_status, ''), COALESCE(image_url, ''), created_at, updated_at FROM users WHERE id > $1 ORDER BY id ASC LIMIT $2;`, (page-1)*pageSize, pageSize)

	for rows.Next() {
		user := new(User)
		rows.Scan(&user.Id, &user.Role, &user.Email, &user.PasswordHash, &user.Username, &user.FullName, &user.DateOfBirth, &user.Gender, &user.PhoneNumber, &user.AccountStatus, &user.ImageUrl, &user.CreatedAt, &user.UpdatedAt)
		users = append(users, user)
	}

	select {
	case count := <-countCh:
		paginatedUsers := &Paginated[*User]{
			TotalItems:  count,
			TotalPages:  uint(math.Ceil(float64(count) / float64(pageSize))),
			CurrentPage: page,
			Items:       users}
		return paginatedUsers, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}

}

func (repo *UserRepository) Create(ctx context.Context, user *UserCore) (uint, error) {
	var id uint
	err := repo.db.QueryRowContext(ctx, `INSERT INTO users(email, password_hash) VALUES($1, $2) RETURNING id;`, user.Email, user.PasswordHash).Scan(&id)
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			switch code := err.Code.Name(); code {
			case "unique_violation":
				return 0, ErrUserAlreadyExists
			default:
				return 0, errors.New(code)
			}
		}
		return 0, err
	}
	return id, err
}

func (repo *UserRepository) DeleteById(ctx context.Context, id uint) error {
	_, err := repo.db.ExecContext(ctx, `DELETE FROM users WHERE id=$1;`, id)
	return err
}

func (repo *UserRepository) Update(ctx context.Context, id uint, updates map[string]any) error {
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
