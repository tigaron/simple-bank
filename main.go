package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"github.com/tigaron/simple-bank/api"
	db "github.com/tigaron/simple-bank/db/sqlc"
)

var (
	dbDriver      = "postgres"
	dbSource      = fmt.Sprintf("postgresql://%s:%s@%s:%s/simple_bank?sslmode=disable", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"))
	serverAddress = "0.0.0.0:8080"
)

func main() {
	conn, err := sql.Open(dbDriver, dbSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	api := api.NewServer(store)

	if err = api.Start(serverAddress); err != nil {
		log.Fatal("cannot start server:", err)
	}
}
