package config

import (
	"fmt"
	"log"
	"os"
)

type EnvVaiable struct {
	Port     string
	DBFile   string
	Password string
}

func СheckEnv() *EnvVaiable {

	var e EnvVaiable
	e.DBFile = "./scheduler.db"
	e.Password = ""
	e.Port = "7540"

	port, ok := os.LookupEnv("TODO_PORT")
	if ok {
		e.Port = port
	}
	dbFile, ok := os.LookupEnv("TODO_DBFILE")
	if ok {
		e.DBFile = dbFile
	}
	password, ok := os.LookupEnv("TODO_PASSWORD")
	if ok {
		e.Password = password
	}
	log.Printf("Значения переменных:\n%s",
		fmt.Sprintf(
			"TODO_PORT: %s\nTODO_DBFILE: %s\nTODO_PASSWORD: %s",
			e.Port,
			e.DBFile,
			e.Password,
		),
	)

	return &e
}
