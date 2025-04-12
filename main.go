package main

import (
	"log"

	"github.com/Jtrx1/go_final_project/config"
	"github.com/Jtrx1/go_final_project/scheduler"
	"github.com/Jtrx1/go_final_project/server"
)

func main() {
	config := config.СheckEnv()
	db, err := scheduler.InitDB(config.DBFile)
	if err != nil {
		log.Println("Ошибка при открытии/инициализации БД: ", err)
	}
	defer db.Close()
	r := server.SetupRouter(db, config.Password)
	err = r.Run(":" + config.Port)
	if err != nil {
		log.Println("Ошибка запуска сервера:", err)
	}
}
