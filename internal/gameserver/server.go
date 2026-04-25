package gameserver

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
	"unicode/utf16"

	"l2c1go/internal/db"
)

type GameServer struct {
	addr string
}

func NewGameServer() *GameServer {
	return &GameServer{addr: ":7777"}
}

func (s *GameServer) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	log.Printf("GameServer listening on %s", s.addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *GameServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	var crypt *GameCrypt
	var accountLogin string
	var char *db.CharData // Состояние персонажа для всей сессии

	for {
		header := make([]byte, 2)
		if _, err := io.ReadFull(conn, header); err != nil {
			break
		}

		length := binary.LittleEndian.Uint16(header) - 2
		if length <= 0 {
			continue
		}

		data := make([]byte, length)
		if _, err := io.ReadFull(conn, data); err != nil {
			break
		}

		if crypt != nil {
			crypt.Decrypt(data)
		}

		packetID := data[0]

		switch packetID {
		case 0x00: // ProtocolVersion
			xorKeyPart := []byte{0x94, 0x35, 0x00, 0x00}
			s.sendFirstKey(conn, xorKeyPart)
			crypt = NewGameCrypt(xorKeyPart)
			log.Printf("GS: Client connected from %s, crypt initialized", conn.RemoteAddr())

		case 0x08, 0x05, 0x02: // AuthLogin
			accountLogin = parseL2String(data[1:])
			if accountLogin == "" {
				accountLogin = "asd"
			}
			log.Printf("GS: Account [%s] authorized", accountLogin)
			
			chars, _ := db.GetCharacters(accountLogin)
			s.sendEncryptedPacket(conn, crypt, PackCharSelectionInfo(accountLogin, chars))

		case 0x0D: // CharacterSelected
			// Читаем индекс выбранного персонажа (0 - первый, 1 - второй и т.д.)
			charIndex := binary.LittleEndian.Uint32(data[1:5])
			
			// Получаем всех персонажей аккаунта
			chars, err := db.GetCharacters(accountLogin)
			if err != nil || int(charIndex) >= len(chars) {
				log.Printf("GS: Ошибка выбора персонажа индекса %d", charIndex)
				break
			}

			// ТЕПЕРЬ МЫ БЕРЕМ ТОГО, КОГО ТЫ ВЫБРАЛ
			char = &chars[charIndex]
			log.Printf("GS: [%s] выбрал персонажа %s (Index: %d)", accountLogin, char.Name, charIndex)

			// Дальше стандартная логика входа...
			inventory, _ := db.GetInventory(char.ObjectID)
			pObj, pItem := getPaperdollArrays(char.ObjectID)

			s.sendEncryptedPacket(conn, crypt, PackSSQInfo())
			s.sendEncryptedPacket(conn, crypt, PackCharSelected(char))
			s.sendEncryptedPacket(conn, crypt, PackQuestList())
			s.sendEncryptedPacket(conn, crypt, PackSkillList())
			
			time.Sleep(100 * time.Millisecond)
			s.sendEncryptedPacket(conn, crypt, PackUserInfo(char, pObj, pItem))
			s.sendEncryptedPacket(conn, crypt, PackItemList(inventory))


		case 0x0F: // RequestItemList от клиента
			if char == nil { break }
			log.Printf("GS: [%s] запрашивает инвентарь", char.Name)
			inventory, _ := db.GetInventory(char.ObjectID)
			// Теперь шлем правильный пакет 0x27
			s.sendEncryptedPacket(conn, crypt, PackItemList(inventory))

		case 0x01: // MoveBackwardToLocation
			if char == nil { break }
			targetX := int32(binary.LittleEndian.Uint32(data[1:5]))
			targetY := int32(binary.LittleEndian.Uint32(data[5:9]))
			targetZ := int32(binary.LittleEndian.Uint32(data[9:13]))
			originX := int32(binary.LittleEndian.Uint32(data[13:17]))
			originY := int32(binary.LittleEndian.Uint32(data[17:21]))
			originZ := int32(binary.LittleEndian.Uint32(data[21:25]))

			log.Printf("GS: [%s] move to %d %d %d", char.Name, targetX, targetY, targetZ)
			
			// ОБНОВЛЯЕМ В БАЗЕ:
			if err := db.UpdateCharacterLocation(char.ObjectID, targetX, targetY, targetZ); err != nil {
				log.Printf("DB Error: %v", err)
			}
            
            // Также обновляем координаты в памяти текущей сессии
            char.X, char.Y, char.Z = targetX, targetY, targetZ

			s.sendEncryptedPacket(conn, crypt, PackCharMoveToLocation(char.ObjectID, targetX, targetY, targetZ, originX, originY, originZ))


		case 0x09: // RequestLogout
			log.Printf("GS: [%s] logout requested", accountLogin)
			// Теперь шлем правильный код подтверждения для C1
			s.sendEncryptedPacket(conn, crypt, []byte{0x96}) 

		case 0x46: // RequestRestart
			log.Printf("GS: [%s] restart requested", accountLogin)
			// Шлем ПРАВИЛЬНЫЙ RestartResponse для твоего клиента
			s.sendEncryptedPacket(conn, crypt, []byte{0x74, 0x01, 0x00, 0x00, 0x00}) 
			
			// В С1 после 0x74 клиент ждет, что его вернут в лобби.
			// Попробуем отправить список персонажей сразу.
			chars, _ := db.GetCharacters(accountLogin)
			s.sendEncryptedPacket(conn, crypt, PackCharSelectionInfo(accountLogin, chars))

		case 0x38: // Say2
			text := parseL2String(data[1:])
			chatType := uint32(0)
			if len(data) >= 4 {
				chatType = binary.LittleEndian.Uint32(data[len(data)-4:])
			}
			// Выводим тип чата в HEX формате (например, 0x08)
			log.Printf("GS: [%s] (ChatType: 0x%02X): %s", char.Name, chatType, text)
			
			s.sendEncryptedPacket(conn, crypt, PackSay2(char.ObjectID, chatType, char.Name, text))

		case 0x63: // RequestQuestList
			s.sendEncryptedPacket(conn, crypt, PackQuestList())

		case 0x03: // RequestEnterWorld
			if char != nil {
				inventory, _ := db.GetInventory(char.ObjectID)
				s.sendEncryptedPacket(conn, crypt, PackItemList(inventory))
				log.Printf("GS: [%s] вошел в мир, отправлено предметов: %d", char.Name, len(inventory))
			}

		case 0x48: // ValidateLocation
			if char == nil { break }
			// Просто ничего не шлем в ответ, но ПРИНИМАЕМ координаты
			if len(data) >= 17 {
				char.X = int32(binary.LittleEndian.Uint32(data[5:9]))
				char.Y = int32(binary.LittleEndian.Uint32(data[9:13]))
				char.Z = int32(binary.LittleEndian.Uint32(data[13:17]))
			}
			// НЕ отправляем PackValidateLocation пока что, 
			// проверим, исчезнет ли мусор в чате без него.


		case 0x0E: // NewCharacter (Нажатие кнопки "Create")
			log.Printf("GS: [%s] открыл экран создания персонажа", accountLogin)
			// Отвечаем, что создавать можно (0x23 в С1)
			s.sendEncryptedPacket(conn, crypt, []byte{0x23, 0x00, 0x00, 0x00, 0x00})

		case 0x0B: // RequestCharacterCreate
			charName := parseL2String(data[1:])
			offset := 1 + len(encodeUTF16(charName))
			
			race := binary.LittleEndian.Uint32(data[offset : offset+4])
			sex := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
			classId := binary.LittleEndian.Uint32(data[offset+8 : offset+12])

			log.Printf("GS: Создание персонажа [%s], Class: %d", charName, classId)

			// Теперь передаем только 5 аргументов, как и просит пакет db
			err := db.CreateCharacter(
				accountLogin, 
				charName, 
				race, 
				classId, 
				sex,
			)
			
			if err != nil {
				log.Printf("GS: Ошибка создания: %v", err)
				s.sendEncryptedPacket(conn, crypt, []byte{0x26, 0x00, 0x00, 0x00, 0x00})
			} else {
				s.sendEncryptedPacket(conn, crypt, []byte{0x25, 0x01, 0x00, 0x00, 0x00})
				chars, _ := db.GetCharacters(accountLogin)
				s.sendEncryptedPacket(conn, crypt, PackCharSelectionInfo(accountLogin, chars))
			}

		case 0x04: // Action (клик по объекту)
			targetID := int32(binary.LittleEndian.Uint32(data[1:5]))
			// Сообщаем клиенту: "Да, ты выделил этот объект"
			s.sendEncryptedPacket(conn, crypt, PackMyTargetSelected(targetID))

		case 0x1B: // SocialAction (Клиент нажал кнопку действия)
			if char == nil {
				break
			}
			// Читаем ID анимации (2 - Поклон, 3 - Махать рукой и т.д.)
			actionID := binary.LittleEndian.Uint32(data[1:5])
			
			log.Printf("GS: [%s] выполняет социальное действие: %d", char.Name, actionID)

			// Отправляем пакет SocialAction обратно клиенту
			// В будущем этот пакет должны получать все игроки вокруг, чтобы видеть твою анимацию
			s.sendEncryptedPacket(conn, crypt, PackSocialAction(char.ObjectID, actionID))

		case 0x37: // RequestTargetCancel (Нажат Esc)
			if char == nil { break }
			s.sendEncryptedPacket(conn, crypt, PackTargetUnselected(char))

		case 0x14: // RequestUseItem
			itemObjID := int32(binary.LittleEndian.Uint32(data[1:5]))
			item, err := db.GetItemByObjID(itemObjID)
			if err != nil {
				log.Printf("GS: Предмет %d не найден в базе", itemObjID)
				break
			}

			// Если это карта (ItemID 1665)
			if item.ItemID == 1665 {
				s.sendEncryptedPacket(conn, crypt, PackShowMiniMap(1665))
				break
			}

			// Определяем индекс слота для куклы (Paperdoll Index)
			var paperdollIdx int32 = -1
			switch item.ItemID {
			case 1:    paperdollIdx = 7  // R_HAND
			case 354:  paperdollIdx = 10 // CHEST
			case 381:  paperdollIdx = 11 // LEGS
			case 2429: paperdollIdx = 9  // GLOVES
			case 2453: paperdollIdx = 12 // FEET
			}

			if paperdollIdx != -1 {
				// Если вещь уже надета — снимаем, иначе надеваем
				if item.Loc == "PAPERDOLL" {
					db.UnquipItem(item.ObjectID)
					log.Printf("GS: [%s] снял предмет %d", char.Name, item.ItemID)
				} else {
					db.EquipItem(char.ObjectID, item.ObjectID, paperdollIdx)
					log.Printf("GS: [%s] надел предмет %d в слот %d", char.Name, item.ItemID, paperdollIdx)
					// Звук успешного надевания (0x2A)
					s.sendEncryptedPacket(conn, crypt, PackEquipItemSuccess(paperdollIdx))
				}

				// --- ПОЛНОЕ ОБНОВЛЕНИЕ КЛИЕНТА ---
				// 1. Получаем актуальное состояние из БД
				pObj, pItem := getPaperdollArrays(char.ObjectID)
				inventory, _ := db.GetInventory(char.ObjectID)

				// 2. Обновляем визуал (модельку)
				s.sendEncryptedPacket(conn, crypt, PackUserInfo(char, pObj, pItem))
				
				// 3. Обновляем инвентарь (рамки вокруг предметов)
				s.sendEncryptedPacket(conn, crypt, PackItemList(inventory))
				
				// 4. Обновляем статы (P.Def, P.Atk) через пакет StatusUpdate
				// Пока отправим базовый, чтобы клиент "проснулся"
				s.sendEncryptedPacket(conn, crypt, PackStatusUpdate(char))
			}

		case 0x11: // RequestUnequipItem
			slot := int32(binary.LittleEndian.Uint32(data[1:5]))
			log.Printf("GS: [%s] запрос на снятие вещи из слота %d", char.Name, slot)
			
			// Нам нужно найти предмет, который надет в этом слоте
			// Добавь функцию GetItemBySlot в db.go
			item, err := db.GetItemBySlot(char.ObjectID, slot)
			if err == nil {
				db.UnquipItem(item.ObjectID)
				
				// Обновляем всё
				pObj, pItem := getPaperdollArrays(char.ObjectID)
				inventory, _ := db.GetInventory(char.ObjectID)
				s.sendEncryptedPacket(conn, crypt, PackUserInfo(char, pObj, pItem))
				s.sendEncryptedPacket(conn, crypt, PackItemList(inventory))
			}

		case 0x57: // RequestShowBoard
			if char == nil { break }
			
			messages, err := db.GetBBSMessages(char.ObjectID)
			if err != nil {
				log.Printf("GS: Ошибка получения BBS: %v", err)
				// Можно отправить пустое окно или системное сообщение
				break
			}
			
			// Формируем контент
			content := "<br><center><font color=\"LEVEL\">Bulletin Board</font></center><br>"
			for _, m := range messages {
				content += fmt.Sprintf("<font color=\"666666\">От: %s</font><br>"+
					"<font color=\"DDDDDD\">Тема: %s</font><br>"+
					"%s<br><br><img src=\"L2UI.SquareGray\" width=280 height=1><br>", 
					m.SenderName, m.Subject, m.Message)
			}

			if len(messages) == 0 {
				content += "<center>Сообщений пока нет.</center>"
			}

			s.sendEncryptedPacket(conn, crypt, PackShowBoard("<html><body>" + content + "</body></html>"))

		default:
			log.Printf("GS: Unknown Packet 0x%02X | Data: %s", packetID, hexDump(data))
		}
	}
}

// ====================== Вспомогательные функции ======================

func parseL2String(data []byte) string {
	var res []uint16
	for i := 0; i+1 < len(data); i += 2 {
		val := binary.LittleEndian.Uint16(data[i : i+2])
		if val == 0 {
			break
		}
		res = append(res, val)
	}
	return string(utf16.Decode(res))
}

func hexDump(data []byte) string {
	var s string
	for _, b := range data {
		s += fmt.Sprintf("%02X ", b)
	}
	return s
}

func (s *GameServer) sendFirstKey(conn net.Conn, key []byte) {
	res := []byte{0x00, 0x01}
	res = append(res, key...)
	res = append(res, []byte{0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}...)
	s.writeRaw(conn, res)
}

func (s *GameServer) sendCharSelectionInfo(conn net.Conn, crypt *GameCrypt, login string) {
	chars, _ := db.GetCharacters(login)
	data := PackCharSelectionInfo(login, chars)
	s.sendEncryptedPacket(conn, crypt, data)
}

func (s *GameServer) sendEncryptedPacket(conn net.Conn, crypt *GameCrypt, data []byte) {
	if crypt == nil {
		s.writeRaw(conn, data)
		return
	}
	work := make([]byte, len(data))
	copy(work, data)

	for len(work)%4 != 0 {
		work = append(work, 0)
	}
	work = AddChecksum(work)

	for len(work)%8 != 0 {
		work = append(work, 0)
	}

	crypt.Encrypt(work)
	s.writeRaw(conn, work)
}

func (s *GameServer) writeRaw(conn net.Conn, data []byte) {
	pkg := make([]byte, len(data)+2)
	binary.LittleEndian.PutUint16(pkg[0:2], uint16(len(data)+2))
	copy(pkg[2:], data)
	_, _ = conn.Write(pkg)
}

func (s *GameServer) handleAdminCommand(conn net.Conn, crypt *GameCrypt, char *db.CharData, text string) {
    switch text {
    case ".whoami":
        msg := fmt.Sprintf("Name: %s, ID: %d, Loc: %d %d %d", char.Name, char.ObjectID, char.X, char.Y, char.Z)
        s.sendEncryptedPacket(conn, crypt, PackSay2(char.ObjectID, 0, "System", msg))
    default:
        s.sendEncryptedPacket(conn, crypt, PackSay2(char.ObjectID, 0, "System", "Unknown command"))
    }
}

func getPaperdollArrays(charID int32) ([15]int32, [15]int32) {
	var objIDs [15]int32
	var itemIDs [15]int32

	// ВАЖНО: убедись, что запрос в db.go или здесь точный
	items, _ := db.GetInventory(charID) 
	for _, it := range items {
		if it.Loc == "PAPERDOLL" { // Проверь регистр!
			idx := it.LocData
			if idx >= 0 && idx < 15 {
				objIDs[idx] = it.ObjectID
				itemIDs[idx] = it.ItemID
			}
		}
	}
	return objIDs, itemIDs
}

func getClassKey(race, classId uint32) string {
	switch race {
	case 0: // Human
		if classId == 0 { return "humanFighter" }
		return "humanMagician"
	case 1: // Elf
		if classId == 18 { return "elfFighter" }
		return "elfMagician"
	case 2: // Dark Elf
		if classId == 31 { return "darkelfFighter" }
		return "darkelfMagician"
	case 3: // Orc
		if classId == 44 { return "orcFighter" }
		return "orcShaman"
	case 4: // Dwarf
		return "dwarfApprentice"
	}
	return "humanFighter"
}
