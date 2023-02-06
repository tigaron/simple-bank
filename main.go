package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
	"github.com/tigaron/simple-bank/api"
	db "github.com/tigaron/simple-bank/db/sqlc"
	"github.com/tigaron/simple-bank/util"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	api := api.NewServer(store)

	if err = api.Start(config.ServerAddress); err != nil {
		log.Fatal("cannot start server:", err)
	}
}
