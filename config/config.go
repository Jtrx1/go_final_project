package config

import "os"

type envVaiable struct {
	Port     string
	DBFile   string
	Password string
}

func checkEnv() *envVaiable  {
	var e envVaiable
	port, ok :=os.LookupEnv("TODO_PORT")
	if ok{
		e.Port=port
	}
	dbFile, ok:=os.LookupEnv("TODO_DBFILE")
	if ok {
		e.DBFile=dbFile
	}
	password, ok:=os.LookupEnv("TODO_PORT")
	if ok {
		e.Password=password
	}
	return &e
	}