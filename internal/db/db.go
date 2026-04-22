// l2c1go/internal/db/db.go
package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Init() {
	dsn := "postgres://postgres:secret@localhost:5432/l2c1?sslmode=disable"
	var err error
	Pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Ошибка пула БД: %v", err)
	}
	if err := Pool.Ping(context.Background()); err != nil {
		log.Fatalf("БД недоступна: %v", err)
	}

	createTables()
	log.Println("База данных PostgreSQL инициализирована")
}

func createTables() {
	ctx := context.Background()

	// DROP TABLE отключён, чтобы персонажи не исчезали при перезапуске
	// Раскомментируй только если хочешь сбросить всё:
	// Pool.Exec(ctx, "DROP TABLE IF EXISTS characters CASCADE;")

	// Аккаунты
	_, _ = Pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS accounts (
		login TEXT PRIMARY KEY,
		password TEXT NOT NULL
	);`)

	// Реестр object_id
	_, _ = Pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS object_id_registry (
		registry_id INT PRIMARY KEY,
		last_object_id INT NOT NULL
	);`)
	_, _ = Pool.Exec(ctx, "INSERT INTO object_id_registry (registry_id, last_object_id) VALUES (1, 100000) ON CONFLICT DO NOTHING")

	// Персонажи
	_, err := Pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS characters (
		object_id INT PRIMARY KEY,
		account_name TEXT REFERENCES accounts(login),
		char_name TEXT UNIQUE NOT NULL,
		race_id INT,
		class_id INT,
		sex INT,
		x INT DEFAULT -70880,
		y INT DEFAULT 257360,
		z INT DEFAULT -3080,
		level INT DEFAULT 1,
		hp DOUBLE PRECISION DEFAULT 100,
		mp DOUBLE PRECISION DEFAULT 50,
		str INT DEFAULT 40,
		dex INT DEFAULT 30,
		con INT DEFAULT 43,
		int INT DEFAULT 21,
		wit INT DEFAULT 11,
		men INT DEFAULT 25
	);`)
	if err != nil {
		log.Fatalf("Ошибка создания таблицы characters: %v", err)
	}

	log.Println("Таблицы проверены/созданы")
}

// CheckAccount — уже была
func CheckAccount(login, password string) (bool, error) {
	var dbPassword string
	err := Pool.QueryRow(context.Background(), "SELECT password FROM accounts WHERE login = $1", login).Scan(&dbPassword)
	if err != nil {
		return false, nil
	}
	return dbPassword == password, nil
}

// === ВАЖНЫЕ ТИП И ФУНКЦИИ ДЛЯ GAMESERVER ===
type CharData struct {
	Name      string
	ObjectID  int32
	Race      int32
	Class     int32
	Sex       int32
	Level     int32
}

func CreateCharacter(login, name string, race, classId, sex uint32) error {
	ctx := context.Background()
	var objID int32

	err := Pool.QueryRow(ctx, `
		UPDATE object_id_registry 
		SET last_object_id = last_object_id + 1 
		WHERE registry_id = 1 
		RETURNING last_object_id
	`).Scan(&objID)
	if err != nil {
		return err
	}

	_, err = Pool.Exec(ctx, `
		INSERT INTO characters 
		(object_id, account_name, char_name, race_id, class_id, sex)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, objID, login, name, race, classId, sex)

	return err
}

func GetCharacters(login string) ([]CharData, error) {
	ctx := context.Background()
	rows, err := Pool.Query(ctx, `
		SELECT char_name, object_id, race_id, class_id, sex, level 
		FROM characters 
		WHERE account_name = $1
	`, login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []CharData
	for rows.Next() {
		var c CharData
		err := rows.Scan(&c.Name, &c.ObjectID, &c.Race, &c.Class, &c.Sex, &c.Level)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		chars = append(chars, c)
	}
	return chars, nil
}

func GetCharacterByName(name string) (*CharData, error) {
	ctx := context.Background()
	var c CharData
	err := Pool.QueryRow(ctx, `
		SELECT char_name, object_id, race_id, class_id, sex, level 
		FROM characters 
		WHERE char_name = $1
	`, name).Scan(&c.Name, &c.ObjectID, &c.Race, &c.Class, &c.Sex, &c.Level)
	
	if err != nil {
		return nil, err
	}
	return &c, nil
}