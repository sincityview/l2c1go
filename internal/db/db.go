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
	ObjectID    int32
	Name        string
	Race, Sex   int32
	ClassID     int32
	X, Y, Z     int32
	Level       int32
	Exp         uint32 // В С1 строго 4 байта
	Sp          int32
	CurHp, MaxHp float64 // В лобби и входе С1 это float64
	CurMp, MaxMp float64
	Karma       int32
	Title       string
	HairStyle, HairColor, Face int32
	Newbie		int32
}

type ItemData struct {
	ObjectID     int32
	ItemID       int32
	Count        int64
	EnchantLevel int32
	Loc          string
	LocData      int32
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
	// Выбираем только те поля, которые ТОЧНО есть в твоей таблице characters
	rows, err := DB.Query(`
		SELECT char_name, charId, race, sex, classid, x, y, z, level, exp, sp, 
			curHp, maxHp, curMp, maxMp, IFNULL(karma, 0), 
			IFNULL(hairStyle, 0), IFNULL(hairColor, 0), IFNULL(face, 0), 
			title, newbie 
		FROM characters WHERE account_name = ?`, login)
	if err != nil {
		log.Printf("DB Query Error: %v", err)
		return nil, err
	}
	defer rows.Close()

	var chars []CharData
	for rows.Next() {
		var c CharData
		var title sql.NullString // Защита от NULL в титуле
		
		// Используем Scan строго по порядку SELECT
		err := rows.Scan(
			&c.Name, &c.ObjectID, &c.Race, &c.Sex, &c.ClassID, 
			&c.X, &c.Y, &c.Z, &c.Level, &c.Exp, &c.Sp, 
			&c.CurHp, &c.MaxHp, &c.CurMp, &c.MaxMp, 
			&c.Karma, &c.HairStyle, &c.HairColor, &c.Face, 
			&title, &c.Newbie,
		)
		
		if err != nil {
			log.Printf("DB Scan Error: %v (Проверь соответствие колонок!)", err)
			continue
		}
		
		c.Title = title.String
		chars = append(chars, c)
	}
	
	log.Printf("DB: Найдено персонажей для [%s]: %d", login, len(chars))
	return chars, nil
}

func getNextObjectID() (int32, error) {
    // Вставляем пустую строку, чтобы получить новый ID
    res, err := DB.Exec("INSERT INTO object_id_registry () VALUES ()")
    if err != nil {
        return 0, err
    }
    id, err := res.LastInsertId()
    return int32(id), err
}

func CreateCharacter(login, name string, race, classId, sex uint32) error {
	tx, err := DB.Begin()
	if err != nil { return err }

	// 1. Получаем шаблон персонажа
	var t struct {
		startX, startY, startZ int32
		hp, mp                 int32
	}
	err = tx.QueryRow(`SELECT startX, startY, startZ, hp, mp 
                       FROM char_templates WHERE classId = ?`, classId).
		Scan(&t.startX, &t.startY, &t.startZ, &t.hp, &t.mp)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Получаем ID для персонажа через новую функцию
	charId, err := getNextObjectID()
	if err != nil {
		tx.Rollback()
		return err
	}

	// 3. Создаем персонажа
	_, err = tx.Exec(`
		INSERT INTO characters (
			charId, account_name, char_name, race, classid, sex, 
			x, y, z, level, curHp, maxHp, curMp, maxMp, 
			sp, exp, newbie, online, title,
			karma, pvpkills, pkkills, hairStyle, hairColor, face)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, 0, 0, 1, 0, '', 0, 0, 0, 0, 0, 0)`,
		charId, login, name, race, classId, sex,
		t.startX, t.startY, t.startZ,
		t.hp, t.hp, t.mp, t.mp)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 4. Вычитываем список вещей в память (чтобы не занимать буфер)
	type StartItem struct {
		id  int32
		amt int32
		eq  int
	}
	var startItems []StartItem
	rows, err := tx.Query("SELECT itemId, amount, equipped FROM char_initial_items WHERE classId = ?", classId)
	if err == nil {
		for rows.Next() {
			var si StartItem
			rows.Scan(&si.id, &si.amt, &si.eq)
			startItems = append(startItems, si)
		}
		rows.Close()
	}

	// 5. Выдаем вещи персонажу
	for _, si := range startItems {
		itemObjID, _ := getNextObjectID() // Используем ту же функцию для ID предметов

		loc := "INVENTORY"
		var locData int32 = 0
		if si.eq == 1 {
			loc = "PAPERDOLL"
			switch si.id {
			case 2369, 6, 2370, 2366, 2371: locData = 7 // Weapon
			case 1146, 425: locData = 10               // Chest
			case 1147, 461: locData = 11               // Legs
			case 2368: locData = 9                     // Gloves
			}
		}

		_, err = tx.Exec(`INSERT INTO items (owner_id, object_id, item_id, count, loc, loc_data) 
                 VALUES (?, ?, ?, ?, ?, ?)`, charId, itemObjID, si.id, si.amt, loc, locData)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func UpdateCharacterLocation(objID int32, x, y, z int32) error {
	_, err := DB.Exec("UPDATE characters SET x = ?, y = ?, z = ? WHERE charId = ?", x, y, z, objID)
	return err
}

func GetInventory(ownerID int32) ([]ItemData, error) {
	// Используем Query, так как предметов много
	rows, err := DB.Query(`
		SELECT object_id, item_id, count, IFNULL(enchant_level, 0), loc, loc_data 
		FROM items 
		WHERE owner_id = ?`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ItemData
	for rows.Next() {
		var it ItemData
		// Важно: порядок сканирования должен строго соответствовать SELECT
		err := rows.Scan(
			&it.ObjectID, 
			&it.ItemID, 
			&it.Count, 
			&it.EnchantLevel, 
			&it.Loc, 
			&it.LocData,
		)
		if err != nil {
			log.Printf("DB: Ошибка чтения предмета: %v", err)
			continue
		}
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
	// Добавляем сортировку ORDER BY server_id ASC
	rows, err := DB.Query("SELECT server_id, host, port FROM gameservers ORDER BY server_id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []GameServerInfo
	for rows.Next() {
		var gs GameServerInfo
		var portStr string
		if err := rows.Scan(&gs.ID, &gs.Host, &portStr); err != nil {
			continue
		}
		fmt.Sscanf(portStr, "%d", &gs.Port)
		servers = append(servers, gs)
	}
	return servers, nil
}

func GetItemByObjID(objID int32) (ItemData, error) {
	var it ItemData
	err := DB.QueryRow(`SELECT object_id, item_id, loc, loc_data FROM items WHERE object_id = ?`, objID).
		Scan(&it.ObjectID, &it.ItemID, &it.Loc, &it.LocData)
	return it, err
}

func EquipItem(charId, itemObjId int32, slot int32) error {
	tx, err := DB.Begin()
	if err != nil { return err }

	// Снимаем только то, что уже надето В ЭТОМ ЖЕ СЛОТЕ
	_, _ = tx.Exec("UPDATE items SET loc = 'INVENTORY', loc_data = 0 WHERE owner_id = ? AND loc = 'PAPERDOLL' AND loc_data = ?", charId, slot)

	// Надеваем новую вещь
	_, err = tx.Exec("UPDATE items SET loc = 'PAPERDOLL', loc_data = ? WHERE object_id = ?", slot, itemObjId)
	
	if err != nil { tx.Rollback(); return err }
	return tx.Commit()
}


func UnquipItem(itemObjId int32) error {
	_, err := DB.Exec("UPDATE items SET loc = 'INVENTORY', loc_data = 0 WHERE object_id = ?", itemObjId)
	return err
}

func GetItemBySlot(charId, slot int32) (ItemData, error) {
	var it ItemData
	err := DB.QueryRow(`SELECT object_id, item_id FROM items WHERE owner_id = ? AND loc = 'PAPERDOLL' AND loc_data = ?`, charId, slot).
		Scan(&it.ObjectID, &it.ItemID)
	return it, err
}

func GetPaperdollForLobby(charId int32) ([15]int32, [15]int32) {
	var objIDs [15]int32
	var itemIDs [15]int32

	rows, err := DB.Query("SELECT object_id, item_id, loc_data FROM items WHERE owner_id = ? AND loc = 'PAPERDOLL'", charId)
	if err != nil { return objIDs, itemIDs }
	defer rows.Close()

	for rows.Next() {
		var objID, itemID, slot int32
		if err := rows.Scan(&objID, &itemID, &slot); err == nil {
			if slot >= 0 && slot < 15 {
				objIDs[slot] = objID
				itemIDs[slot] = itemID
			}
		}
	}
	return objIDs, itemIDs
}
