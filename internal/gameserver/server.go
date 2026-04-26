package gameserver

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
	"unicode/utf16"

	"darkages/internal/db"
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
			charIndex := binary.LittleEndian.Uint32(data[1:5])
			chars, err := db.GetCharacters(accountLogin)
			if err != nil || int(charIndex) >= len(chars) {
				log.Printf("GS: Ошибка выбора персонажа индекса %d", charIndex)
				break
			}

			char = &chars[charIndex]
			log.Printf("GS: [%s] выбрал персонажа %s (Index: %d)", accountLogin, char.Name, charIndex)

			// 1. Сначала подтверждаем выбор персонажа
			s.sendEncryptedPacket(conn, crypt, PackCharSelected(char))
			
			// 2. Обязательная пауза для С1, чтобы клиент переключил состояние экрана
			time.Sleep(300 * time.Millisecond)

			// 3. Базовая информация о мире
			s.sendEncryptedPacket(conn, crypt, PackSSQInfo())
			s.sendEncryptedPacket(conn, crypt, PackQuestList())
			s.sendEncryptedPacket(conn, crypt, PackSkillList())
            
            // В С1 после этого клиент обычно шлет пакет 0x03 (EnterWorld)

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
			if char == nil { break }
			log.Printf("GS: [%s] запрашивает вход в мир (EnterWorld)", char.Name)

			// Получаем актуальные данные
			inventory, _ := db.GetInventory(char.ObjectID)
			pObj, pItem := getPaperdollArrays(char.ObjectID)

			// 1. UserInfo — самый важный пакет для спавна
			s.sendEncryptedPacket(conn, crypt, PackUserInfo(char, pObj, pItem))
			
			// 2. Список вещей
			time.Sleep(100 * time.Millisecond)
			s.sendEncryptedPacket(conn, crypt, PackItemList(inventory))
            
            // 3. Системное приветствие (опционально)
			// Вместо PackSystemMessage(34)
			welcomeText := "Welcome to Dark Ages!"
			s.sendEncryptedPacket(conn, crypt, PackSay2(0, 0x00, "Server", welcomeText))


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
				break
			}

			// 1. Определяем тип предмета через его ID
			bodyPart := getBodyPartByID(item.ItemID) 
			slot := getSlotByBodyPart(bodyPart)

			// 2. Если это экипировка
			if slot != -1 {
				if item.Loc == "PAPERDOLL" {
					db.UnquipItem(item.ObjectID)
				} else {
					db.EquipItem(char.ObjectID, item.ObjectID, slot)
					s.sendEncryptedPacket(conn, crypt, PackEquipItemSuccess(slot))
				}

				// 3. Обновляем состояние клиента
				pObj, pItem := getPaperdollArrays(char.ObjectID)
				inventory, _ := db.GetInventory(char.ObjectID)
				s.sendEncryptedPacket(conn, crypt, PackUserInfo(char, pObj, pItem))
				s.sendEncryptedPacket(conn, crypt, PackItemList(inventory))
				s.sendEncryptedPacket(conn, crypt, PackStatusUpdate(char))
			} else if item.ItemID == 1665 { // Карта
				s.sendEncryptedPacket(conn, crypt, PackShowMiniMap(1665))
			}

		case 0x11: // RequestUnequipItem
			bodyPart := binary.LittleEndian.Uint32(data[1:5])
			slot := getSlotByBodyPart(bodyPart) // Конвертируем маску в индекс 0-14
			
			if slot != -1 {
				item, err := db.GetItemBySlot(char.ObjectID, slot)
				if err == nil {
					db.UnquipItem(item.ObjectID)
					log.Printf("GS: [%s] снял предмет %d из слота %d", char.Name, item.ItemID, slot)
					
					// СРАЗУ ОБНОВЛЯЕМ КЛИЕНТ
					pObj, pItem := getPaperdollArrays(char.ObjectID)
					inventory, _ := db.GetInventory(char.ObjectID)
					s.sendEncryptedPacket(conn, crypt, PackUserInfo(char, pObj, pItem))
					s.sendEncryptedPacket(conn, crypt, PackItemList(inventory))
				}
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
	// Исправлено: 07x00 заменено на 0x00
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

func getSlotByBodyPart(bodyPart uint32) int32 {
	switch bodyPart {
	case 0x80:   return 7  // R_HAND
	case 0x400:  return 10 // CHEST
	case 0x800:  return 11 // LEGS
	case 0x200:  return 9  // GLOVES
	case 0x1000: return 12 // FEET
	case 0x08:   return 1  // REAR
	case 0x10:   return 2  // LEAR
	case 0x20:   return 3  // NECK
	case 0x40:   return 4  // RFINGER (Правое кольцо)
	case 0x04:   return 6  // Head/Helmet
	}
	return -1
}

