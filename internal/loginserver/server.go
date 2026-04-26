// darkages/internal/loginserver/server.go
package loginserver

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"

	"darkages/internal/db"
)

type Server struct {
	addr        string
	protocolRev uint32
	token       []byte
}

func NewServer() *Server {
	return &Server{
		addr:        ":2106",
		protocolRev: 419, // C1
		token:       []byte{0x5b, 0x3b, 0x27, 0x2e, 0x5d, 0x39, 0x34, 0x2d, 0x33, 0x31, 0x3d, 0x3d, 0x2d, 0x25, 0x26, 0x40, 0x21, 0x5e, 0x2b, 0x5d},
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("LoginServer listening on %s (Protocol %d)", s.addr, s.protocolRev)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	crypt := NewCrypt()
	var accountLogin string
	clientIP, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	if err := s.sendInit(conn); err != nil {
		log.Printf("Failed to send Init packet: %v", err)
		return
	}

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

		crypt.Decrypt(data)
		packetID := data[0]

		switch packetID {
		case 0x00, 0x58: 
			login := "asd"
			pass := "dsa" 

			ok, err := db.CheckAccount(login, pass, clientIP)
			if err != nil {
				log.Printf("LS: DB Error: %v", err)
				s.sendLoginFail(conn, crypt, 0x01) 
				break
			}

			if ok {
				accountLogin = login
				log.Printf("LS: Account [%s] authorized from IP: %s", login, clientIP)
				s.sendLoginOk(conn, crypt, 0x11223344, 0x55667788) 
			} else {
				log.Printf("LS: Login failed for [%s]", login)
				s.sendLoginFail(conn, crypt, 0x01)
			}

		case 0x05, 0xA4: // RequestServerList
			if accountLogin == "" {
				s.sendLoginFail(conn, crypt, 0x01)
				return
			}
			s.sendServerList(conn, crypt)

		case 0x02, 0xCB, 0x4A: // RequestServerLogin
			if accountLogin == "" {
				s.sendLoginFail(conn, crypt, 0x01)
				return
			}
			s.sendPlayOk(conn, crypt, 0x11223344, 0x55667788)

		case 0x0B: // RequestCharacterCreate
			charName := ""
			for i := 1; i < 33 && i < len(data); i += 2 {
				if data[i] == 0 { break }
				charName += string(data[i])
			}

			if charName == "" || accountLogin == "" {
				s.sendCharacterCreateFail(conn, crypt)
				continue
			}

			race := binary.LittleEndian.Uint32(data[33:37])
			gender := binary.LittleEndian.Uint32(data[37:41])
			classId := binary.LittleEndian.Uint32(data[41:45])

			// ИСПРАВЛЕНО: Теперь ровно 5 аргументов, как в твоем db.go
			if err := db.CreateCharacter(accountLogin, charName, race, classId, gender); err != nil {
				log.Printf("DB Error: %v", err)
				s.sendCharacterCreateFail(conn, crypt)
				continue
			}

			s.sendCharacterCreateSuccess(conn, crypt)
			s.sendCharSelectionInfo(conn, crypt, accountLogin)

		default:
			log.Printf("Unknown Packet: 0x%02x", packetID)
		}
	}
}

// ==================== Packet Helpers (добавлено сюда) ====================

// encodeUTF16 преобразует строку в L2 Unicode (UTF-16LE + null terminator)
func encodeUTF16(s string) []byte {
	res := make([]byte, 0, len(s)*2+2)
	for _, r := range s {
		buf := make([]byte, 2)
		binary.LittleEndian.PutUint16(buf, uint16(r))
		res = append(res, buf...)
	}
	res = append(res, 0x00, 0x00)
	return res
}

// PackCharSelectionInfo — теперь используем версию из gameserver (самая полная)
func PackCharSelectionInfo(login string, chars []db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x1F)

	if len(chars) == 0 {
		binary.Write(buf, binary.LittleEndian, uint32(0))
		return buf.Bytes()
	}

	binary.Write(buf, binary.LittleEndian, uint32(len(chars)))

	for _, char := range chars {
		// ПОЛУЧАЕМ ПРЕДМЕТЫ ДЛЯ ОТОБРАЖЕНИЯ В ЛОББИ (как в gameserver)
		pObj, pItem := db.GetPaperdollForLobby(char.ObjectID)

		buf.Write(encodeUTF16(char.Name))
		binary.Write(buf, binary.LittleEndian, uint32(char.ObjectID))
		buf.Write(encodeUTF16(login))
		binary.Write(buf, binary.LittleEndian, uint32(0x55555555)) // SessionID
		binary.Write(buf, binary.LittleEndian, uint32(0))          // ClanID
		binary.Write(buf, binary.LittleEndian, uint32(0))          // Placeholder

		binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
		binary.Write(buf, binary.LittleEndian, uint32(char.Race))
		binary.Write(buf, binary.LittleEndian, uint32(char.ClassID))
		binary.Write(buf, binary.LittleEndian, uint32(1)) // Active

		binary.Write(buf, binary.LittleEndian, int32(char.X))
		binary.Write(buf, binary.LittleEndian, int32(char.Y))
		binary.Write(buf, binary.LittleEndian, int32(char.Z))

		binary.Write(buf, binary.LittleEndian, float64(char.CurHp))
		binary.Write(buf, binary.LittleEndian, float64(char.CurMp))

		binary.Write(buf, binary.LittleEndian, uint32(char.Sp))
		binary.Write(buf, binary.LittleEndian, uint32(char.Exp))
		binary.Write(buf, binary.LittleEndian, uint32(char.Level))
		binary.Write(buf, binary.LittleEndian, uint32(char.Karma))

		// 9 Reserved
		for i := 0; i < 9; i++ {
			binary.Write(buf, binary.LittleEndian, uint32(0))
		}

		// Paperdoll ObjectIDs
		for i := 0; i < 15; i++ {
			binary.Write(buf, binary.LittleEndian, uint32(pObj[i]))
		}

		// Paperdoll ItemIDs
		for i := 0; i < 15; i++ {
			binary.Write(buf, binary.LittleEndian, uint32(pItem[i]))
		}

		binary.Write(buf, binary.LittleEndian, uint32(char.HairStyle))
		binary.Write(buf, binary.LittleEndian, uint32(char.HairColor))
		binary.Write(buf, binary.LittleEndian, uint32(char.Face))

		binary.Write(buf, binary.LittleEndian, float64(char.MaxHp))
		binary.Write(buf, binary.LittleEndian, float64(char.MaxMp))

		binary.Write(buf, binary.LittleEndian, uint32(0)) // delete flag
	}
	return buf.Bytes()
}

// ==================== Вспомогательные методы ====================

func (s *Server) sendInit(conn net.Conn) error {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x00)
	binary.Write(buf, binary.LittleEndian, uint32(0x12345678))
	binary.Write(buf, binary.LittleEndian, s.protocolRev)

	rsaModule := make([]byte, 128)
	buf.Write(rsaModule)

	buf.Write(make([]byte, 16))
	buf.Write(s.token)
	buf.WriteByte(0x00)

	return s.writeRaw(conn, buf.Bytes())
}

func (s *Server) sendLoginOk(conn net.Conn, crypt *Crypt, k1, k2 uint32) {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x03)
	binary.Write(buf, binary.LittleEndian, k1)
	binary.Write(buf, binary.LittleEndian, k2)
	buf.Write(make([]byte, 8))
	s.sendEncryptedPacket(conn, crypt, buf.Bytes())
}

func (s *Server) sendLoginFail(conn net.Conn, crypt *Crypt, reason byte) {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x01)
	buf.WriteByte(reason)
	s.sendEncryptedPacket(conn, crypt, buf.Bytes())
}

func (s *Server) sendServerList(conn net.Conn, crypt *Crypt) {
	servers, _ := db.GetGameServers()

	buf := new(bytes.Buffer)
	buf.WriteByte(0x04)                 // Opcode
	buf.WriteByte(uint8(len(servers)))  // Кол-во серверов
	buf.WriteByte(0x01)                 // Last Server ID

	for _, gs := range servers {
		buf.WriteByte(uint8(gs.ID))     // 1 = Bartz, 2 = Sieghardt...
		
		ip := net.ParseIP(gs.Host).To4()
		if ip == nil { ip = []byte{127, 0, 0, 1} }
		buf.Write(ip)                   // IP (4 байта)

		binary.Write(buf, binary.LittleEndian, int32(gs.Port)) // Порт (4 байта)
		buf.WriteByte(0x00)             // Age Limit
		buf.WriteByte(0x00)             // PK Free (0/1)
		
		binary.Write(buf, binary.LittleEndian, uint16(0))   // Текущий онлайн
		binary.Write(buf, binary.LittleEndian, uint16(100)) // Макс. онлайн
		
		// ВНИМАНИЕ: В С1 статус (Up/Down) часто идет ПОСЛЕ онлайна
		buf.WriteByte(0x01)             // Status (1 - Up)

		// КРИТИЧНО: Бит-маска типа сервера (4 байта)
		// 0x01: Normal, 0x02: Relax, 0x04: Public Test, 0x08: No Label...
		binary.Write(buf, binary.LittleEndian, uint32(0x00)) 

		buf.WriteByte(0x00)             // Brackets (0 - No, 1 - Yes)
	}

	s.sendEncryptedPacket(conn, crypt, buf.Bytes())
}


func (s *Server) sendPlayOk(conn net.Conn, crypt *Crypt, k1, k2 uint32) {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x07)
	binary.Write(buf, binary.LittleEndian, k1)
	binary.Write(buf, binary.LittleEndian, k2)
	buf.WriteByte(0x00)
	s.sendEncryptedPacket(conn, crypt, buf.Bytes())
}

func (s *Server) sendCharacterCreateSuccess(conn net.Conn, crypt *Crypt) {
	data := []byte{0x25, 0x01, 0x00, 0x00, 0x00}
	s.sendEncryptedPacket(conn, crypt, data)
}

func (s *Server) sendCharacterCreateFail(conn net.Conn, crypt *Crypt) {
	data := []byte{0x25, 0x00, 0x00, 0x00, 0x00}
	s.sendEncryptedPacket(conn, crypt, data)
}

func (s *Server) sendCharSelectionInfo(conn net.Conn, crypt *Crypt, login string) {
	chars, err := db.GetCharacters(login)
	if err != nil {
		log.Printf("Failed to get characters for %s: %v", login, err)
		chars = nil
	}
	data := PackCharSelectionInfo(login, chars)
	s.sendEncryptedPacket(conn, crypt, data)
}

func (s *Server) sendEncryptedPacket(conn net.Conn, crypt *Crypt, data []byte) {
	padLen := 8 - (len(data) % 8)
	if padLen < 8 {
		data = append(data, make([]byte, padLen)...)
	}
	crypt.Encrypt(data)
	_ = s.writeRaw(conn, data)
}

func (s *Server) writeRaw(conn net.Conn, data []byte) error {
	out := make([]byte, len(data)+2)
	binary.LittleEndian.PutUint16(out[0:2], uint16(len(data)+2))
	copy(out[2:], data)
	_, err := conn.Write(out)
	return err
}