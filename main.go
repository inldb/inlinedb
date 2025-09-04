package main

import (
	"fmt"
	"log"

	"github.com/MultiX0/db-test/api"
	dbclass "github.com/MultiX0/db-test/db"
)

func main() {
	defer fmt.Println("End...")

	fmt.Println("Starting...")

	err := dbclass.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	server := api.NewAPIServer(":1212")
	server.Run()

}
