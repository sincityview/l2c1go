// darkages/cmd/loginserver/main.go
package main

import (
	"log"
	"darkages/internal/db"
	"darkages/internal/loginserver"
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