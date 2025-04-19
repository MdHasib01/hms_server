package main

import (
	"log"

	"github.com/MdHasib01/hms_server/internal/db"
	"github.com/MdHasib01/hms_server/internal/env"
	"github.com/MdHasib01/hms_server/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/hmsDB?sslmode=disable")

	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	store := store.NewStorage(conn)
	db.Seed(store)
}
