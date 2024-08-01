package repo

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	_ "modernc.org/sqlite"

	"github.com/rohitxdev/go-api-template/config"
	"github.com/rohitxdev/go-api-template/util"
)

/*----------------------------------- Paginated Response Type ----------------------------------- */

type Paginated[T any] struct {
	TotalItems  uint `json:"total_items"`
	TotalPages  uint `json:"total_pages"`
	CurrentPage uint `json:"current_page"`
	Items       []T  `json:"items"`
}

/*----------------------------------- PostgreSQL Connection ----------------------------------- */

var PostgresDB = func() *sql.DB {
	db, err := sql.Open("postgres", config.DB_URL)
	if err != nil {
		panic("could not connect to PostgreSQL database: " + err.Error())
	}
	util.RegisterCleanUp("postgres connection", func() error { return db.Close() })
	return db
}()

/*-----------------------------------SQLite Connection ----------------------------------- */

var SQLiteDB = func() *sql.DB {
	db, err := sql.Open("sqlite", "sqlite.db")
	if err != nil {
		panic("could not open SQLite database: " + err.Error())
	}
	util.RegisterCleanUp("sqlite connection", func() error { return db.Close() })
	return db
}()

/*----------------------------------- MongoDB Connection ----------------------------------- */

var MongoDBClient = func() *mongo.Client {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(config.MONGODB_URL))
	if err != nil {
		panic("could not connect to mongodb: " + err.Error())
	}
	util.RegisterCleanUp("mongodb connection", func() error {
		return client.Disconnect(context.TODO())
	})
	return client
}()

/*----------------------------------- Redis Connection ----------------------------------- */

var RedisClient = func() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     config.REDIS_HOST + ":" + config.REDIS_PORT,
		Username: config.REDIS_USERNAME,
		Password: config.REDIS_PASSWORD,
	})
	util.RegisterCleanUp("redis connection", func() error { return client.Close() })
	return client
}()
