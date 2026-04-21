// l2c1go/cmd/loginserver
package main

import (
	"log"
	"l2c1go/internal/db"
	"l2c1go/internal/loginserver"
)

func main() {
	// 1. Инициализируем базу данных
	db.Init()

	// 2. Запускаем сервер
	log.Println("Starting LoginServer...")
	server := loginserver.NewServer()
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start LoginServer: %v", err)
	}
}