// l2c1go/internal/loginserver
package loginserver

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"

	"l2c1go/internal/db" // Импорт твоего пакета БД
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

	// 1. Отправляем Init пакет (0x00)
	if err := s.sendInit(conn); err != nil {
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

		// Дешифруем пакет
		crypt.Decrypt(data)
		packetID := data[0]

		switch packetID {
		case 0x00: // RequestAuthLogin
			// В C1 логин и пароль идут в фиксированных полях (14 и 16 байт)
			login := string(bytes.TrimRight(data[1:15], "\x00"))
			password := string(bytes.TrimRight(data[15:31], "\x00"))

			// --- ПРОВЕРКА ЧЕРЕЗ БАЗУ ДАННЫХ ---
			isValid, err := db.CheckAccount(login, password)
			if err != nil {
				log.Printf("DB Error for %s: %v", login, err)
				return
			}
			if !isValid {
				log.Printf("Auth Failed: login [%s] or password incorrect", login)
				// В идеале тут слать LoginFail (0x01), но пока просто рвем коннект
				return
			}

			log.Printf("User [%s] authenticated successfully via DB", login)
			
			// Сессионные ключи для гейм-сервера
			s.sendLoginOk(conn, crypt, 0x11223344, 0x55667788)

		case 0x05: // RequestServerList
			s.sendServerList(conn, crypt)

		case 0x02: // RequestServerLogin
			s.sendPlayOk(conn, crypt, 0x11223344, 0x55667788)

		case 0x0B: // RequestCharacterCreate
			// ... парсинг имени уже есть ...
			race := binary.LittleEndian.Uint32(data[33:37])
			gender := binary.LittleEndian.Uint32(data[37:41]) // Следующие 4 байта
			classId := binary.LittleEndian.Uint32(data[41:45])
			
			log.Printf("GS: Создание персонажа [%s], Пол: %d", charName, gender)
			db.CreateCharacter(accountLogin, charName, race, classId, gender)
			
			s.sendCharacterCreateSuccess(conn, crypt)
			s.sendCharSelectionInfo(conn, crypt, accountLogin)

		default:
			log.Printf("Unknown LS Packet: 0x%02x", packetID)
		}
	}
}

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

func (s *Server) sendServerList(conn net.Conn, crypt *Crypt) {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x04)
	buf.WriteByte(0x01) // Count
	buf.WriteByte(0x00) // Last ID

	buf.WriteByte(0x01)                            // Server ID
	buf.Write([]byte{127, 0, 0, 1})                // IP
	binary.Write(buf, binary.LittleEndian, uint32(7777))
	buf.WriteByte(0x00) // Age
	buf.WriteByte(0x00) // pvp
	binary.Write(buf, binary.LittleEndian, uint16(0))
	binary.Write(buf, binary.LittleEndian, uint16(1000))
	buf.WriteByte(0x01) // Status

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

func (s *Server) sendEncryptedPacket(conn net.Conn, crypt *Crypt, data []byte) {
	// Дополнение до кратности 8 для Blowfish
	padLen := 8 - (len(data) % 8)
	if padLen < 8 {
		data = append(data, make([]byte, padLen)...)
	}
	crypt.Encrypt(data)
	s.writeRaw(conn, data)
}

func (s *Server) writeRaw(conn net.Conn, data []byte) error {
	out := make([]byte, len(data)+2)
	binary.LittleEndian.PutUint16(out[0:2], uint16(len(data)+2))
	copy(out[2:], data)
	_, err := conn.Write(out)
	return err
}
