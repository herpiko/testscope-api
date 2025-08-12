package main

import (
	"log"

	"github.com/joho/godotenv"
)

var app App

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	_ = godotenv.Load()

	app = App{}
	app.Init()
	app.Run("0.0.0.0:8000")
}
