package main

import "github.com/famusovsky/AvitoTestTask/internal/usersegmentation"

func main() {
	app := usersegmentation.Create()
	app.Run()
}
