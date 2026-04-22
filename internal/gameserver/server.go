// l2c1go/internal/gameserver/server.go
package gameserver

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"

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
			log.Printf("GS: Client connected from %s", conn.RemoteAddr())

		case 0x08, 0x05, 0x02: // AuthLogin
			accountLogin = ""
			for i := 1; i < len(data) && i < 30; i += 2 {
				if data[i] == 0 {
					break
				}
				accountLogin += string(data[i])
			}
			if accountLogin == "" {
				log.Printf("GS: Empty login received")
				break
			}
			log.Printf("GS: Игрок [%s] вошел", accountLogin)
			s.sendCharSelectionInfo(conn, crypt, accountLogin)

		case 0x0E: // NewCharacter
			s.sendNewCharacterSuccess(conn, crypt)

		case 0x0B: // RequestCharacterCreate
			log.Printf("GS: [0x0B] RAW HEX: %s", hexDump(data))

			charName := ""
			i := 1
			for i+1 < len(data) {
				if data[i] == 0 && data[i+1] == 0 {
					break
				}
				if data[i] != 0 {
					charName += string(data[i])
				}
				i += 2
			}

			if charName == "" || accountLogin == "" {
				log.Printf("GS: Invalid character creation request")
				break
			}

			// Временно захардкодим Human Female Warrior
			race := uint32(0)   // Human
			sex := uint32(1)    // Female
			classId := uint32(0) // Fighter

			log.Printf("GS: Создание персонажа [%s] → Human Female Warrior", charName)

			if err := db.CreateCharacter(accountLogin, charName, race, classId, sex); err != nil {
				log.Printf("DB Error: %v", err)
			} else {
				log.Printf("GS: Персонаж [%s] создан", charName)
				s.sendCharacterCreateSuccess(conn, crypt)
				s.sendCharSelectionInfo(conn, crypt, accountLogin)
			}

		case 0x0D: // Start Game (выбор персонажа)
			// Улучшенный парсинг имени
			selectedCharName := ""
			for i := 1; i+1 < len(data); i += 2 {
				if data[i] == 0 && data[i+1] == 0 {
					break
				}
				if data[i] != 0 {
					selectedCharName += string(data[i])
				}
			}
			if selectedCharName == "" {
				selectedCharName = "Unknown"
			}

			log.Printf("GS: Игрок [%s] выбрал персонажа [%s] → вход в мир", accountLogin, selectedCharName)

			s.sendCharSelected(conn, crypt)
			s.sendSSQInfo(conn, crypt)
			s.sendQuestList(conn, crypt)
			s.sendSkillList(conn, crypt)

			time.Sleep(150 * time.Millisecond)
			s.sendUserInfo(conn, crypt, selectedCharName)

			time.Sleep(100 * time.Millisecond)
			s.sendItemList(conn, crypt)

			log.Printf("GS: Персонаж [%s] вошёл в мир", selectedCharName)

		case 0x66: // Likely RequestSkillList or similar
			log.Printf("GS: Received packet 0x66 from [%s]", accountLogin)
			// Можно отправить пустой SkillList повторно, если нужно
			s.sendSkillList(conn, crypt)

		case 0x68: // Likely RequestHenna or another info packet
			log.Printf("GS: Received packet 0x68 from [%s]", accountLogin)
			// Пока игнорируем, клиент часто спамит этим при входе

		case 0xD3: // Pong
			log.Printf("GS: Pong from [%s]", accountLogin)

		default:
			log.Printf("GS: Unknown packet 0x%02x from [%s]", packetID, accountLogin)
		}
	}
}

// ====================== Вспомогательные функции ======================

func hexDump(data []byte) string {
	var s string
	for _, b := range data {
		s += fmt.Sprintf("%02X ", b)
	}
	return s
}

func (s *GameServer) sendFirstKey(conn net.Conn, keyPart []byte) {
	res := []byte{0x00, 0x01}
	res = append(res, keyPart...)
	res = append(res, []byte{0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00}...)
	s.writeRaw(conn, res)
}

func (s *GameServer) sendCharSelectionInfo(conn net.Conn, crypt *GameCrypt, login string) {
	chars, err := db.GetCharacters(login)
	if err != nil {
		log.Printf("Failed to get characters: %v", err)
		chars = nil
	}
	data := PackCharSelectionInfo(login, chars)
	s.sendEncryptedPacket(conn, crypt, data)
}

func (s *GameServer) sendNewCharacterSuccess(conn net.Conn, crypt *GameCrypt) {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x23)
	binary.Write(buf, binary.LittleEndian, uint32(0))
	s.sendEncryptedPacket(conn, crypt, buf.Bytes())
}

func (s *GameServer) sendCharacterCreateSuccess(conn net.Conn, crypt *GameCrypt) {
	data := []byte{0x25, 0x01, 0x00, 0x00, 0x00}
	s.sendEncryptedPacket(conn, crypt, data)
}

func (s *GameServer) sendCharSelected(conn net.Conn, crypt *GameCrypt) {
	s.sendEncryptedPacket(conn, crypt, PackCharSelected("assa"))
}

func (s *GameServer) sendSSQInfo(conn net.Conn, crypt *GameCrypt) {
	data := []byte{0xF8, 0x01, 0x02}
	s.sendEncryptedPacket(conn, crypt, data)
}

func (s *GameServer) sendQuestList(conn net.Conn, crypt *GameCrypt) {
	s.sendEncryptedPacket(conn, crypt, PackQuestList())
}

func (s *GameServer) sendSkillList(conn net.Conn, crypt *GameCrypt) {
	s.sendEncryptedPacket(conn, crypt, PackSkillList())
}

func (s *GameServer) sendUserInfo(conn net.Conn, crypt *GameCrypt, charName string) {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x04) // UserInfo

	// Основная информация
	buf.Write(encodeUTF16(charName))
	binary.Write(buf, binary.LittleEndian, uint32(10001)) // objectId
	buf.Write(encodeUTF16("Human Fighter"))               // title

	binary.Write(buf, binary.LittleEndian, uint32(0))     // race = Human
	binary.Write(buf, binary.LittleEndian, uint32(1))     // sex = Female
	binary.Write(buf, binary.LittleEndian, uint32(0))     // classId = Fighter (Human Warrior)

	binary.Write(buf, binary.LittleEndian, uint32(1))     // level

	// Координаты (проверенные для C1)
	binary.Write(buf, binary.LittleEndian, int32(-70880))
	binary.Write(buf, binary.LittleEndian, int32(257360))
	binary.Write(buf, binary.LittleEndian, int32(-3080))

	binary.Write(buf, binary.LittleEndian, float64(126.0)) // HP
	binary.Write(buf, binary.LittleEndian, float64(38.0))  // MP

	// Статы из шаблона
	binary.Write(buf, binary.LittleEndian, uint32(40)) // STR
	binary.Write(buf, binary.LittleEndian, uint32(30)) // DEX
	binary.Write(buf, binary.LittleEndian, uint32(43)) // CON
	binary.Write(buf, binary.LittleEndian, uint32(21)) // INT
	binary.Write(buf, binary.LittleEndian, uint32(11)) // WIT
	binary.Write(buf, binary.LittleEndian, uint32(25)) // MEN

	// Базовые боевые характеристики
	binary.Write(buf, binary.LittleEndian, uint32(4))   // pAtk
	binary.Write(buf, binary.LittleEndian, uint32(72))  // pDef
	binary.Write(buf, binary.LittleEndian, uint32(3))   // mAtk
	binary.Write(buf, binary.LittleEndian, uint32(47))  // mDef

	// Скорости
	binary.Write(buf, binary.LittleEndian, uint32(330)) // pSpd
	binary.Write(buf, binary.LittleEndian, uint32(213)) // mSpd

	// Заполняем остаток полей нулями (очень важно для стабильности)
	for i := 0; i < 55; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(0))
	}

	// Heading
	binary.Write(buf, binary.LittleEndian, uint32(0))

	s.sendEncryptedPacket(conn, crypt, buf.Bytes())
}

func (s *GameServer) sendItemList(conn net.Conn, crypt *GameCrypt) {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x1B)
	buf.WriteByte(0x00)
	binary.Write(buf, binary.LittleEndian, uint16(0))
	s.sendEncryptedPacket(conn, crypt, buf.Bytes())
}

func (s *GameServer) sendEncryptedPacket(conn net.Conn, crypt *GameCrypt, data []byte) {
	if crypt == nil {
		s.writeRaw(conn, data)
		return
	}

	packet := make([]byte, len(data))
	copy(packet, data)

	for len(packet)%4 != 0 {
		packet = append(packet, 0x00)
	}
	packet = AddChecksum(packet)

	for len(packet)%8 != 0 {
		packet = append(packet, 0x00)
	}

	crypt.Encrypt(packet)
	s.writeRaw(conn, packet)
}

func (s *GameServer) writeRaw(conn net.Conn, data []byte) {
	pkg := make([]byte, len(data)+2)
	binary.LittleEndian.PutUint16(pkg[0:2], uint16(len(data)+2))
	copy(pkg[2:], data)
	_, _ = conn.Write(pkg)
}