package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// CharData — структура, полностью соответствующая таблице Mobius
type CharData struct {
	AccountName string
	ObjectID    int32  // charId
	Name        string // char_name
	Level       int32
	MaxHp       int32
	CurHp       int32
	MaxMp       int32
	CurMp       int32
	Face        int32
	HairStyle   int32
	HairColor   int32
	Sex         int32
	X, Y, Z     int32
	Exp         int64
	Sp          int64
	Karma       int32
	ClassID     int32 

// classid
	Race        int32
	Title       string
}

type ItemData struct {
	ObjectID     int32
	ItemID       int32
	Count        int64
	EnchantLevel int32
	Loc          string
	LocData      int32 // Для PAPERDOLL здесь хранится ID слота (экипировка)
}

func Init() {
	// user:password@tcp(localhost:3306)/darkages
	dsn := "mariadb-user:mariadb-password@tcp(localhost:3306)/darkages"
	
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Ошибка открытия MariaDB: %v", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(10)
	DB.SetConnMaxLifetime(time.Hour)

	if err := DB.Ping(); err != nil {
		log.Fatalf("БД недоступна: %v", err)
	}

	createTables()
	log.Println("MariaDB инициализирована успешно")
}

func createTables() {
	// Реестр для выдачи новых ID персонажам
	_, err := DB.Exec(`CREATE TABLE IF NOT EXISTS object_id_registry (
		registry_id INT PRIMARY KEY,
		last_object_id INT NOT NULL
	);`)
	if err != nil { log.Fatal(err) }

	_, _ = DB.Exec("INSERT IGNORE INTO object_id_registry (registry_id, last_object_id) VALUES (1, 100000)")
	
	// Таблицу characters и accounts мы НЕ создаем здесь, так как ты импортировал их из Mobius SQL.
	// Но если их нет, Mobius-логика может сломаться.
}

func CheckAccount(login, password string) (bool, error) {
	var dbPassword sql.NullString
	err := DB.QueryRow("SELECT password FROM accounts WHERE login = ?", login).Scan(&dbPassword)
	
	if err == sql.ErrNoRows {
		log.Printf("Авторегистрация аккаунта: %s", login)
		_, err = DB.Exec("INSERT INTO accounts (login, password) VALUES (?, ?)", login, password)
		return true, err
	}
	
	if err != nil { return false, err }
	return dbPassword.String == password, nil
}

func GetCharacters(login string) ([]CharData, error) {
	// Тянем основные поля для лобби и входа
	rows, err := DB.Query(`
		SELECT char_name, charId, race, classid, sex, level, x, y, z, title, curHp, maxHp, curMp, maxMp 
		FROM characters 
		WHERE account_name = ?
	`, login)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chars []CharData
	for rows.Next() {
		var c CharData
		var title sql.NullString
		// Сканируем в соответствии с порядком в SELECT
		err := rows.Scan(&c.Name, &c.ObjectID, &c.Race, &c.ClassID, &c.Sex, &c.Level, &c.X, &c.Y, &c.Z, &title, &c.CurHp, &c.MaxHp, &c.CurMp, &c.MaxMp)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		c.Title = title.String
		chars = append(chars, c)
	}
	return chars, nil
}

func CreateCharacter(login, name string, race, classId, sex uint32) error {
	tx, err := DB.Begin()
	if err != nil { return err }

	var objID int32
	err = tx.QueryRow("SELECT last_object_id FROM object_id_registry WHERE registry_id = 1 FOR UPDATE").Scan(&objID)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	objID++
	
	_, err = tx.Exec("UPDATE object_id_registry SET last_object_id = ? WHERE registry_id = 1", objID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Вставка в таблицу Mobius (charId, char_name и т.д.)
	_, err = tx.Exec(`
		INSERT INTO characters (
			charId, account_name, char_name, race, classid, sex, 
			x, y, z, curHp, maxHp, curMp, maxMp, level, title
		)
		VALUES (?, ?, ?, ?, ?, ?, -70880, 257360, -3080, 100, 100, 50, 50, 1, '')
	`, objID, login, name, race, classId, sex)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func UpdateCharacterLocation(objID int32, x, y, z int32) error {
	_, err := DB.Exec("UPDATE characters SET x = ?, y = ?, z = ? WHERE charId = ?", x, y, z, objID)
	return err
}

func GetInventory(ownerID int32) ([]ItemData, error) {
	rows, err := DB.Query(`
		SELECT object_id, item_id, count, enchant_level, loc, loc_data 
		FROM items 
		WHERE owner_id = ?
	`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ItemData
	for rows.Next() {
		var it ItemData
		err := rows.Scan(&it.ObjectID, &it.ItemID, &it.Count, &it.EnchantLevel, &it.Loc, &it.LocData)
		if err != nil {
			continue
		}
		items = append(items, it)
	}
	return items, nil
}
