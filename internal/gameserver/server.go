// l2c1go/internal/gameserver
package gameserver

import (
	"bytes"
	"encoding/binary"
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
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *GameServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	var crypt *GameCrypt
	var accountLogin string
	done := make(chan bool)
	defer close(done)

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

		// ВАЖНО: берем ID как первый байт массива
		packetID := data[0]

		switch packetID {
		case 0x00: // ProtocolVersion
			xorKeyPart := []byte{0x94, 0x35, 0x00, 0x00}
			s.sendFirstKey(conn, xorKeyPart)
			crypt = NewGameCrypt(xorKeyPart)
			go func() {
				ticker := time.NewTicker(60 * time.Second)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						s.sendNetPing(conn, crypt)
					case <-done:
						return
					}
				}
			}()

		case 0x08, 0x05, 0x02: // AuthLogin
			accountLogin = ""
			for i := 1; i < len(data) && i < 30; i += 2 {
				if data[i] == 0 { break }
				accountLogin += string(data[i])
			}
			log.Printf("GS: Игрок [%s] вошел", accountLogin)
			s.sendCharSelectionInfo(conn, crypt, accountLogin)

		case 0x0E: // NewCharacter
			s.sendNewCharacterSuccess(conn, crypt)

		case 0x0B: // RequestCharacterCreate
			// Достаем имя (первые 32 байта после ID)
			charName := ""
			for i := 1; i < 33; i += 2 {
				if data[i] == 0 { break }
				charName += string(data[i])
			}

			// Читаем параметры по смещениям (каждый D = 4 байта)
			// data[33-37] - race, data[37-41] - gender
			race := binary.LittleEndian.Uint32(data[33:37])
			gender := binary.LittleEndian.Uint32(data[37:41])
			classId := binary.LittleEndian.Uint32(data[41:45])
			
			log.Printf("GS: Создание персонажа [%s] Пол: %d Раса: %d", charName, gender, race)
			
			if err := db.CreateCharacter(accountLogin, charName, race, classId, gender); err != nil {
				log.Printf("DB Error: %v", err)
			} else {
				s.sendCharacterCreateSuccess(conn, crypt)
				s.sendCharSelectionInfo(conn, crypt, accountLogin)
			}

		case 0x0D: // Start Game
			s.sendCharSelected(conn, crypt)
			s.sendQuestList(conn, crypt)
			s.sendSkillList(conn, crypt)
			s.sendEnterWorld(conn, crypt)

		case 0xD3: // Pong
			log.Printf("GS: Pong from %s", accountLogin)
		}
	}
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
	buf := new(bytes.Buffer)
	buf.WriteByte(0x23)
	binary.Write(buf, binary.LittleEndian, uint32(0))
	s.sendEncryptedPacket(conn, crypt, buf.Bytes())
}

func (s *GameServer) sendCharacterCreateSuccess(conn net.Conn, crypt *GameCrypt) {
	data := []byte{0x25, 0x01, 0x00, 0x00, 0x00}
	s.sendEncryptedPacket(conn, crypt, data)
}

func (s *GameServer) sendNetPing(conn net.Conn, crypt *GameCrypt) {
	s.sendEncryptedPacket(conn, crypt, []byte{0xD3, 0x00, 0x00, 0x00, 0x00})
}

func (s *GameServer) sendQuestList(conn net.Conn, crypt *GameCrypt) {
	s.sendEncryptedPacket(conn, crypt, PackQuestList())
}

func (s *GameServer) sendSkillList(conn net.Conn, crypt *GameCrypt) {
	s.sendEncryptedPacket(conn, crypt, PackSkillList())
}

func (s *GameServer) sendCharSelected(conn net.Conn, crypt *GameCrypt) {
	s.sendEncryptedPacket(conn, crypt, PackCharSelected("assa"))
}

func (s *GameServer) sendEnterWorld(conn net.Conn, crypt *GameCrypt) {
	s.sendEncryptedPacket(conn, crypt, PackEnterWorld())
}

func (s *GameServer) sendEncryptedPacket(conn net.Conn, crypt *GameCrypt, data []byte) {
	// 1. Сначала дополняем пакет нулями, чтобы его длина была кратна 4 байтам
	// Это нужно, чтобы AddChecksum не падал
	for len(data)%4 != 0 {
		data = append(data, 0x00)
	}

	// 2. Теперь считаем и добавляем чексумму (она добавит еще 4 байта)
	data = AddChecksum(data)

	// 3. Теперь дополняем до кратности 8 байт для XOR шифрования
	for len(data)%8 != 0 {
		data = append(data, 0x00)
	}

	// 4. Шифруем
	crypt.Encrypt(data)

	// 5. Отправляем в сокет
	s.writeRaw(conn, data)
}


func (s *GameServer) writeRaw(conn net.Conn, data []byte) {
	pkg := make([]byte, len(data)+2)
	binary.LittleEndian.PutUint16(pkg[0:2], uint16(len(data)+2))
	copy(pkg[2:], data)
	conn.Write(pkg)
}
