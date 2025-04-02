package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"frappuccino/internal/server"
	"frappuccino/internal/utils"

	_ "github.com/lib/pq"
)

func main() {
	config, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Postgres.Host,
		config.Postgres.Port,
		config.Postgres.User,
		config.Postgres.Password,
		config.Postgres.Database,
	)

	db, err := connectDB(connString)
	if err != nil {
		for i := range 5 {
			time.Sleep(5 * time.Second)
			log.Printf("Retrying to connect to db (%v)", i+1)
			db, err = connectDB(connString) // retry connecting to db
			if err == nil {
				break
			} else if i == 4 { // after 5 retries give up
				log.Fatal(err)
			}
		}
	}
	defer db.Close()

	server := server.NewServer(":8080", db, utils.GetLogger())
	server.RunServer()
}

func connectDB(connectionString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	log.Println("Connected to Database successfully!")

	return db, nil
}
