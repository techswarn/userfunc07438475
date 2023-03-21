package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
)

func Main(args map[string]interface{}) map[string]interface{} {
	log.Logger = log.Output(os.Stdout)
	log.Info().Msg("Received a message event")
	var (
		dbHost = os.Getenv("dbHost")
		dbPort = os.Getenv("dbPort")
		dbUser = os.Getenv("dbUser")
		dbPass = os.Getenv("dbPass")
		dbName = os.Getenv("dbName")
	)
	connInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	db, err := sql.Open("postgres", connInfo)
	if err != nil {
		log.Error().Err(err).Msg("Error opening database")
		return map[string]interface{}{
			"statusCode": http.StatusInternalServerError,
			"body":       err.Error(),
		}
	}
	defer db.Close()
	log.Info().Msg("Successfully connected to database")

	return map[string]interface{}{
		"statusCode": http.StatusCreated,
		"body":       "Message sent",
	}
}
