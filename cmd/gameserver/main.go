// l2c1go/cmd/gameserver
package main

import (
	"darkages/internal/gameserver"
	"log"
	"darkages/internal/db"
)


func main() {
	// 1. Инициализируем базу данных
	db.Init()

	// 2. Запускаем сервер
	log.Println("Starting GameServer...")
	server := gameserver.NewGameServer()
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start GameServer: %v", err)
	}
}

