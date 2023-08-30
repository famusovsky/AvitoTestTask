package main

import (
	"flag"
	"log"
	"os"

	_ "github.com/famusovsky/AvitoTestTask/docs"
	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation"
	"github.com/famusovsky/AvitoTestTask/internal/usersegmentation/postgres"
	"github.com/famusovsky/AvitoTestTask/pkg/db"
	_ "github.com/lib/pq"
)

// XXX DO NOT FORGET ABOUT COMMENTS
// TODO normal logging
// TODO change getUsers from body to query params

// @title User Segmentation API
// @description This is a User Segmentation API server, made for Avito Backend Trainee Assignment 2023.
func main() {
	addr := flag.String("addr", ":8080", "HTTP address")
	createTables := flag.Bool("create_tables", false, "Create tables in database")
	flag.Parse()

	// TODO change logger
	logger := log.New(os.Stdout, "LOG\t", log.Ldate|log.Ltime)

	db, err := db.OpenViaEnvVars("postgres")
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	dbProcessor, err := postgres.GetModel(db, *createTables)
	if err != nil {
		logger.Fatal(err)
	}

	app := usersegmentation.CreateApp(logger, dbProcessor)

	app.Run(*addr)
}
