package repo

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/rohitxdev/go-api-template/internal/env"
)

var Db = connectToDb(env.DB_URL)

func connectToDb(URL string) *sql.DB {
	log.Println("Connecting to database...")
	db, err := sql.Open("postgres", URL)
	if err != nil {
		log.Fatalln("Error when connecting to database:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalln("Error when pinging database:", err)
	}
	log.Println("Connected to database successfully ✅")
	return db
}
