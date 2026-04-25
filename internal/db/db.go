package db

import (
	"database/sql"
	"log"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

type BBSMessage struct {
	SenderName string
	Subject    string
	Message    string
}

type GameServerInfo struct {
	ID   int
	Host string
	Port int
}

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
	ClassID     int32 // classid
	Race        int32
	Title       string
}

type ItemData struct {
	ObjectID     int32 // object_id
	ItemID       int32 // item_id
	Count        int64
	EnchantLevel int32 // enchant_level
	Loc          string
	LocData      int32 // loc_data
}

func Init() {
    dsn := "mariadb-user:mariadb-password@tcp(localhost:3306)/darkages"
    var err error
    DB, err = sql.Open("mysql", dsn)
    if err != nil { log.Fatal(err) }

    if err := DB.Ping(); err != nil { log.Fatal(err) }
    
    // Эту таблицу можно оставить, так как она чисто техническая для выдачи ID
    _, _ = DB.Exec(`CREATE TABLE IF NOT EXISTS object_id_registry (registry_id INT PRIMARY KEY, last_object_id INT NOT NULL);`)
    _, _ = DB.Exec("INSERT IGNORE INTO object_id_registry (registry_id, last_object_id) VALUES (1, 100000)")
    
	log.Println("MariaDB инициализирована успешно")
}

func CheckAccount(login, password string, ip string) (bool, error) {
	var dbPassword sql.NullString
	err := DB.QueryRow("SELECT password FROM accounts WHERE login = ?", login).Scan(&dbPassword)
	
	if err == sql.ErrNoRows {
		// Авторегистрация с SHA256
		_, err = DB.Exec(`INSERT INTO accounts (login, password, lastIP, lastactive) 
						 VALUES (?, SHA2(?, 256), ?, CURRENT_TIMESTAMP)`, login, password, ip)
		return true, err
	}
	
	if err != nil { return false, err }

	var isValid bool
	err = DB.QueryRow("SELECT (password = SHA2(?, 256)) FROM accounts WHERE login = ?", password, login).Scan(&isValid)
	if err != nil { return false, err }

	if isValid {
		_, _ = DB.Exec("UPDATE accounts SET lastIP = ?, lastactive = CURRENT_TIMESTAMP WHERE login = ?", ip, login)
		return true, nil
	}
	
	return false, nil
}

func GetCharacters(login string) ([]CharData, error) {
	rows, err := DB.Query(`
		SELECT char_name, charId, race, classid, sex, level, x, y, z, title, curHp, maxHp, curMp, maxMp 
		FROM characters 
		WHERE account_name = ?`, login)
	if err != nil { return nil, err }
	defer rows.Close()

	var chars []CharData
	for rows.Next() {
		var c CharData
		var title sql.NullString
		err := rows.Scan(&c.Name, &c.ObjectID, &c.Race, &c.ClassID, &c.Sex, &c.Level, &c.X, &c.Y, &c.Z, &title, &c.CurHp, &c.MaxHp, &c.CurMp, &c.MaxMp)
		if err != nil { continue }
		c.Title = title.String
		chars = append(chars, c)
	}
	return chars, nil
}

func CreateCharacter(login, name string, race, classId, sex uint32, x, y, z int32, items []string) error {
	tx, err := DB.Begin()
	if err != nil { return err }

	var charId int32
	err = tx.QueryRow("SELECT last_object_id FROM object_id_registry WHERE registry_id = 1 FOR UPDATE").Scan(&charId)
	if err != nil { tx.Rollback(); return err }
	charId++
	tx.Exec("UPDATE object_id_registry SET last_object_id = ? WHERE registry_id = 1", charId)

	_, err = tx.Exec(`
		INSERT INTO characters (charId, account_name, char_name, race, classid, sex, x, y, z, curHp, maxHp, curMp, maxMp, level, title)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 100, 100, 50, 50, 1, '')`, 
		charId, login, name, race, classId, sex, x, y, z)
	if err != nil { tx.Rollback(); return err }

	// Цикл выдачи предметов (GiveItem) используй как раньше, но с колонками:
	// (owner_id, object_id, item_id, count, loc, loc_data)
	
	return tx.Commit()
}

func UpdateCharacterLocation(objID int32, x, y, z int32) error {
	_, err := DB.Exec("UPDATE characters SET x = ?, y = ?, z = ? WHERE charId = ?", x, y, z, objID)
	return err
}

func GetInventory(ownerID int32) ([]ItemData, error) {
	rows, err := DB.Query(`SELECT object_id, item_id, count, enchant_level, loc, loc_data FROM items WHERE owner_id = ?`, ownerID)
	if err != nil { return nil, err }
	defer rows.Close()

	var items []ItemData
	for rows.Next() {
		var it ItemData
		err := rows.Scan(&it.ObjectID, &it.ItemID, &it.Count, &it.EnchantLevel, &it.Loc, &it.LocData)
		if err != nil { continue }
		items = append(items, it)
	}
	return items, nil
}

func GetBBSMessages(charId int32) ([]BBSMessage, error) {
	// Ищем сообщения, где получатель либо 0 (все), либо текущий персонаж
	rows, err := DB.Query(`
		SELECT sender_id, subject, message 
		FROM bbs 
		WHERE receiver_id = 0 OR receiver_id = ? 
		ORDER BY id DESC LIMIT 5`, charId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []BBSMessage
	for rows.Next() {
		var msg BBSMessage
		var senderId int32
		if err := rows.Scan(&senderId, &msg.Subject, &msg.Message); err != nil {
			continue
		}
		
		// Логика отправителя: если 0, подставляем "Admin"
		msg.SenderName = "Admin"
		if senderId != 0 {
			// В будущем тут можно сделать SELECT name FROM characters WHERE charId = senderId
			msg.SenderName = fmt.Sprintf("User_%d", senderId)
		}
		
		messages = append(messages, msg)
	}
	return messages, nil
}

func GetGameServers() ([]GameServerInfo, error) {
	rows, err := DB.Query("SELECT server_id, host, port FROM gameservers")
	if err != nil { return nil, err }
	defer rows.Close()

	var servers []GameServerInfo
	for rows.Next() {
		var gs GameServerInfo
		var portStr string // читаем порт как строку
		if err := rows.Scan(&gs.ID, &gs.Host, &portStr); err != nil {
			continue
		}
		// Конвертируем строку "7777" в число 7777
		fmt.Sscanf(portStr, "%d", &gs.Port)
		servers = append(servers, gs)
	}
	return servers, nil
}
