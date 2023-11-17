package repo

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/rohitxdev/go-api-template/internal/env"
)

/*----------------------------------- PostgreSQL Connection ----------------------------------- */

var postgresDb = connectToPostgres(env.DB_URL)

func connectToPostgres(URL string) *sql.DB {
	log.Println("Connecting to PostgreSQL database...")

	db, err := sql.Open("postgres", URL)
	if err != nil {
		panic(err)
	}

	log.Println("Connected to PostgreSQL database successfully ✅")

	return db
}

/*----------------------------------- Redis Connection ----------------------------------- */

var Redis = connectToRedis()

func connectToRedis() *redis.Client {
	log.Println("Connecting to redis...")

	rds := redis.NewClient(&redis.Options{
		Addr:     env.REDIS_HOST + ":" + env.REDIS_PORT,
		Username: env.REDIS_USERNAME,
		Password: env.REDIS_PASSWORD})

	log.Println("Connected to redis successfully ✅")

	return rds
}

/*----------------------------------- Paginated Response Type ----------------------------------- */

type Paginated[T any] struct {
	TotalItems  uint `json:"total_items"`
	TotalPages  uint `json:"total_pages"`
	CurrentPage uint `json:"current_page"`
	Items       []T  `json:"items"`
}
