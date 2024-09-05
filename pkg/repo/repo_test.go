package repo_test

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/rohitxdev/go-api-template/pkg/assert"
	"github.com/rohitxdev/go-api-template/pkg/repo"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNew(t *testing.T) {
	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name string
		args args
		want *repo.Repo
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := repo.New(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRepo_Migrate(t *testing.T) {
	tests := []struct {
		name    string
		repo    *repo.Repo
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.repo.Migrate(); (err != nil) != tt.wantErr {
				t.Errorf("Repo.Migrate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRepo(t *testing.T) {
	ctx := context.Background()

	// Set up the PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpassword",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(60 * time.Second),
	}

	postgresC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer postgresC.Terminate(ctx)

	// Get the container host and port
	host, err := postgresC.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}

	port, err := postgresC.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatal(err)
	}

	dsn := fmt.Sprintf("postgres://testuser:testpassword@%s:%s/testdb?sslmode=disable", host, port.Port())

	// Connect to PostgreSQL using pgx
	db, err := sql.Open("postgres", dsn)
	assert.Nil(t, err)
	defer db.Close()

	r := repo.New(db)
	err = r.Migrate()
	assert.Nil(t, err)

	t.Run("Create user", func(t *testing.T) {
		user := repo.UserCore{
			Email:        "test@test.com",
			PasswordHash: "testpassword",
		}
		id, err := r.CreateUser(ctx, &user)
		assert.Nil(t, err)
		assert.NotEqual(t, id, "")
	})
}
