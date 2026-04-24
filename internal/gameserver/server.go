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

		case 0x0D: // Character Selected (Нажата кнопка START)
			chars, _ := db.GetCharacters(accountLogin)
			if len(chars) > 0 {
				char = &chars[0] // Сохраняем ссылку на персонажа
				log.Printf("GS: Character [%s] (ID: %d) entering world...", char.Name, char.ObjectID)

				s.sendEncryptedPacket(conn, crypt, PackSSQInfo())
				s.sendEncryptedPacket(conn, crypt, PackCharSelected(char))
				s.sendEncryptedPacket(conn, crypt, PackQuestList())
				s.sendEncryptedPacket(conn, crypt, PackSkillList())

				time.Sleep(200 * time.Millisecond)
				
				s.sendEncryptedPacket(conn, crypt, PackUserInfo(char))
				s.sendEncryptedPacket(conn, crypt, PackItemList())
				
				log.Printf("GS: All entry packets sent for [%s].", char.Name)
			}

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
			chatType := binary.LittleEndian.Uint32(data[len(data)-4:])

			// Если сообщение начинается с точки — это команда
			if len(text) > 0 && text[0] == '.' {
				s.handleAdminCommand(conn, crypt, char, text)
			} else {
				log.Printf("GS: [%s]: %s", char.Name, text)
				s.sendEncryptedPacket(conn, crypt, PackSay2(char.ObjectID, chatType, char.Name, text))
			}

		case 0x63: // RequestQuestList
			s.sendEncryptedPacket(conn, crypt, PackQuestList())

		case 0x03: // RequestEnterWorld
			// Попробуем отправить ID 34 (Welcome). 
			// Если снова кританет, значит в твоем протоколе 0x7A имеет другую длину.
			s.sendEncryptedPacket(conn, crypt, PackSystemMessage(34))

		case 0x48: // ValidatePosition
			// Клиент просто сообщает свои координаты для синхронизации. 
			// Пока ничего не делаем, просто чтобы не падало в default.

		case 0x0E: // NewCharacter (Нажатие кнопки "Create")
			log.Printf("GS: [%s] открыл экран создания персонажа", accountLogin)
			// Отвечаем, что создавать можно (0x23 в С1)
			s.sendEncryptedPacket(conn, crypt, []byte{0x23, 0x00, 0x00, 0x00, 0x00})

		case 0x0B: // RequestCharacterCreate (Нажата кнопка "OK" в создании)
			charName := parseL2String(data[1:])
			
			// Смещение: ищем параметры после имени
			nameBytes := encodeUTF16(charName)
			offset := 1 + len(nameBytes)
			
			if len(data) >= offset+12 {
				race := binary.LittleEndian.Uint32(data[offset : offset+4])
				sex := binary.LittleEndian.Uint32(data[offset+4 : offset+8])
				classId := binary.LittleEndian.Uint32(data[offset+8 : offset+12])

				log.Printf("GS: Создание чара [%s] Race:%d Sex:%d Class:%d", charName, race, sex, classId)

				err := db.CreateCharacter(accountLogin, charName, race, classId, sex)
				if err != nil {
					log.Printf("GS: DB Error: %v", err)
					s.sendEncryptedPacket(conn, crypt, []byte{0x25, 0x00, 0x00, 0x00, 0x00}) // Fail
				} else {
					log.Printf("GS: Персонаж [%s] успешно создан", charName)
					s.sendEncryptedPacket(conn, crypt, []byte{0x25, 0x01, 0x00, 0x00, 0x00}) // Success
					
					// Обновляем список персонажей для клиента
					chars, _ := db.GetCharacters(accountLogin)
					s.sendEncryptedPacket(conn, crypt, PackCharSelectionInfo(accountLogin, chars))
				}
			}

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
