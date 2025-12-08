package main

import (
	root_config "catalogizer/config"
	"catalogizer/database"
	"context"
	"fmt"
	"log"
)

func main() {
	db, err := database.NewConnection(&root_config.DatabaseConfig{
		Path: "./data/catalog.db",
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := db.RunMigrations(context.Background()); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Migrations completed successfully")
}
