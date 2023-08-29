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

// @title User Segmentation API
// @description This is a User Segmentation API server, made for Avito Backend Trainee Assignment 2023.
func main() {
	addr := flag.String("addr", ":8080", "HTTP address")
	flag.Parse()

	// TODO change logger
	errorLog := log.New(os.Stdout, "ERR\t", log.Ldate|log.Ltime)

	db, err := db.OpenViaEnvVars()
	if err != nil {
		errorLog.Fatal(err)
	}

	dbProcessor, err := postgres.GetModel(db)
	if err != nil {
		errorLog.Fatal(err)
	}

	app := usersegmentation.CreateApp(errorLog, dbProcessor)

	app.Run(*addr)
}
