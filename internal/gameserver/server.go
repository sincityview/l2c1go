package gameserver

import (
	"encoding/binary"
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
			// В С1 ключ часто статический для инициализации
			xorKeyPart := []byte{0x94, 0x35, 0x00, 0x00}
			s.sendFirstKey(conn, xorKeyPart)
			crypt = NewGameCrypt(xorKeyPart)
			log.Printf("GS: Client connected, crypt initialized")

		case 0x08, 0x05, 0x02: // AuthLogin (вход на GS после выбора сервера)
			// Парсим логин, чтобы знать, чьих персонажей грузить
			accountLogin = parseL2String(data[1:])
			if accountLogin == "" {
				accountLogin = "asd" // Фоллбэк для тестов
			}
			log.Printf("GS: Account [%s] authorized on GameServer", accountLogin)
			
			// Отправляем список персонажей
			chars, _ := db.GetCharacters(accountLogin)
			s.sendEncryptedPacket(conn, crypt, PackCharSelectionInfo(accountLogin, chars))

		case 0x0E: // NewCharacter (переход к экрану создания)
			s.sendEncryptedPacket(conn, crypt, []byte{0x23, 0x00, 0x00, 0x00, 0x00})

		case 0x0B: // RequestCharacterCreate (создание персонажа)
			charName := parseL2String(data[1:])
			// Смещения могут плавать, для теста берем фиксированные статы
			if err := db.CreateCharacter(accountLogin, charName, 0, 0, 0); err != nil {
				log.Printf("GS: DB Error creating char: %v", err)
			} else {
				log.Printf("GS: Character [%s] created", charName)
				s.sendEncryptedPacket(conn, crypt, []byte{0x25, 0x01, 0x00, 0x00, 0x00})
				chars, _ := db.GetCharacters(accountLogin)
				s.sendEncryptedPacket(conn, crypt, PackCharSelectionInfo(accountLogin, chars))
			}

		case 0x0D: // Character Selected (Нажата кнопка START)
			// 1. Ищем персонажа в базе
			chars, _ := db.GetCharacters(accountLogin)
			if len(chars) > 0 {
				char := &chars[0] // Берем первого персонажа аккаунта для теста
				log.Printf("GS: Character [%s] (ID: %d) entering world...", char.Name, char.ObjectID)

				// 2. СТРОГИЙ ПОРЯДОК ИЗ ТВОЕГО ОПИСАНИЯ:
				// 18) SSQInfo (SignsSky)
				s.sendEncryptedPacket(conn, crypt, PackSSQInfo())
				
				// 19) CharacterSelected (наш 0x15)
				s.sendEncryptedPacket(conn, crypt, PackCharSelected(char))
				
				// Дополнительные пакеты состояния
				s.sendEncryptedPacket(conn, crypt, PackQuestList())
				s.sendEncryptedPacket(conn, crypt, PackSkillList())

				// 3. Пауза, чтобы клиент успел прогрузить меши
				time.Sleep(200 * time.Millisecond)
				
				// 4. UserInfo — финальный аккорд для появления модельки
				s.sendEncryptedPacket(conn, crypt, PackUserInfo(char))
				
				log.Printf("GS: UserInfo sent. Character should appear.")
			} else {
				log.Printf("GS: No characters found for [%s] to enter world", accountLogin)
			}

		case 0xD3: // Pong
			// Клиент подтверждает, что он жив
		default:
			log.Printf("GS: Unknown packet 0x%02x from %s", packetID, accountLogin)
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

func (s *GameServer) sendFirstKey(conn net.Conn, keyPart []byte) {
	res := []byte{0x00, 0x01}
	res = append(res, keyPart...)
	res = append(res, []byte{0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}...)
	s.writeRaw(conn, res)
}

func (s *GameServer) sendCharSelectionInfo(conn net.Conn, crypt *GameCrypt, login string) {
	chars, _ := db.GetCharacters(login)
	data := PackCharSelectionInfo(login, chars)
	s.sendEncryptedPacket(conn, crypt, data)
}

func (s *GameServer) sendNewCharacterSuccess(conn net.Conn, crypt *GameCrypt) {
	s.sendEncryptedPacket(conn, crypt, []byte{0x23, 0x00, 0x00, 0x00, 0x00})
}

func (s *GameServer) sendCharacterCreateSuccess(conn net.Conn, crypt *GameCrypt) {
	s.sendEncryptedPacket(conn, crypt, []byte{0x25, 0x01, 0x00, 0x00, 0x00})
}

func (s *GameServer) sendEncryptedPacket(conn net.Conn, crypt *GameCrypt, data []byte) {
	if crypt == nil {
		s.writeRaw(conn, data)
		return
	}
	// Работаем с копией, чтобы не портить исходные данные
	work := make([]byte, len(data))
	copy(work, data)

	// Padding для чексуммы (4 байта)
	for len(work)%4 != 0 {
		work = append(work, 0)
	}
	work = AddChecksum(work)

	// Padding для Blowfish/XOR (8 байт)
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
