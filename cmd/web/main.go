package main

import (
	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation"
	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/postgres"
	"github.com/famusovsky/AvitoTestTask/pkg/db"
	_ "github.com/lib/pq"
)

func main() {
	db, err := db.OpenViaEnvVars()
	if err != nil {
		panic(err)
	}

	app, err := usersegmentation.CreateApp(db, postgres.GetModel)
	if err != nil {
		panic(err)
	}

	app.Run()
}
