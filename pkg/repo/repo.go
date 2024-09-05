package repo

import (
	"database/sql"
)

type Repo struct {
	db    *sql.DB
	stmts *Stmts
}

func New(db *sql.DB) *Repo {
	stmts, err := prepareStmts(db)
	if err != nil {
		panic(err)
	}
	return &Repo{
		db:    db,
		stmts: stmts,
	}
}

func (repo *Repo) Migrate() error {
	if _, err := repo.db.Exec("CREATE EXTENSION IF NOT EXISTS CITEXT;"); err != nil {
		return err
	}
	_, err := repo.stmts.CreateUserTable.Exec()
	return err
}

type Stmts struct {
	CreateUserTable *sql.Stmt
	GetUserById     *sql.Stmt
	GetUserByEmail  *sql.Stmt
	CreateUser      *sql.Stmt
	DeleteUserById  *sql.Stmt
	Update          *sql.Stmt
}

func prepareStmts(db *sql.DB) (*Stmts, error) {
	stmts := Stmts{}

	createUserTable, err := db.Prepare(`CREATE TABLE IF NOT EXISTS users(
    id TEXT PRIMARY KEY,
	role TEXT CHECK (role IN ('user', 'admin')) DEFAULT 'user',
    email CITEXT NOT NULL UNIQUE CHECK (LENGTH(email)<=64),
    password_hash TEXT NOT NULL CHECK (LENGTH(email)<=72),
	username TEXT UNIQUE CHECK (LENGTH(username)<=32) DEFAULT '',
    full_name TEXT CHECK (LENGTH(full_name)<=64) DEFAULT '',
    date_of_birth DATE,
    gender TEXT CHECK (gender IN ('male', 'female', 'other')),
	phone_number TEXT CHECK (LENGTH(phone_number)<=16),
	account_status TEXT CHECK (account_status IN ('active', 'suspended', 'banned')) DEFAULT 'active',
	image_url TEXT,
    created_at TIMESTAMPTZ DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ DEFAULT current_timestamp
);`)
	if err != nil {
		return nil, err
	}
	stmts.CreateUserTable = createUserTable

	getUserById, err := db.Prepare(`SELECT id, role, email, password_hash, COALESCE(username, ''), COALESCE(full_name, ''), COALESCE(date_of_birth, '-infinity'), COALESCE(gender, ''), COALESCE(phone_number, ''), COALESCE(account_status, ''), COALESCE(image_url, ''), created_at, updated_at FROM users WHERE id=$1 LIMIT 1;`)
	if err != nil {
		return nil, err
	}
	stmts.GetUserById = getUserById

	getUserByEmail, err := db.Prepare(`SELECT id, role, email, password_hash, COALESCE(username, ''), COALESCE(full_name, ''), COALESCE(date_of_birth, '-infinity'), COALESCE(gender, ''), COALESCE(phone_number, ''), COALESCE(account_status, ''), COALESCE(image_url, ''), created_at, updated_at FROM users WHERE email=$1 LIMIT 1;`)
	if err != nil {
		return nil, err
	}
	stmts.GetUserByEmail = getUserByEmail

	createUser, err := db.Prepare(`INSERT INTO users(email, password_hash) VALUES($1, $2) RETURNING id;`)
	if err != nil {
		return nil, err
	}
	stmts.CreateUser = createUser

	deleteUserById, err := db.Prepare(`DELETE FROM users WHERE id=$1;`)
	if err != nil {
		return nil, err
	}
	stmts.DeleteUserById = deleteUserById

	return &stmts, nil
}
