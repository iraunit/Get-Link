package main

import (
	"fmt"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading Env", err)
	}
	app := InitializeApp()
	app.Start()
}
