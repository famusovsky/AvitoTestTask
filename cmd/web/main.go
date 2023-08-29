package main

import (
	"log"

	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation"
	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/postgres"
	"github.com/famusovsky/AvitoTestTask/pkg/db"
	_ "github.com/lib/pq"
)

// XXX DO NOT FORGET ABOUT COMMENTS

func main() {
	db, err := db.OpenViaEnvVars()
	if err != nil {
		panic(err)
	}

	dbProcessor, err := postgres.GetModel(db)
	if err != nil {
		panic(err)
	}

	// TODO change logger
	app := usersegmentation.CreateApp(log.Default(), dbProcessor)

	app.Run()
}
