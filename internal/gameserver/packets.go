// l2c1go/internal/gameserver/packets.go
package gameserver

import (
	"bytes"
	"encoding/binary"

	"l2c1go/internal/db"
)

// encodeUTF16 преобразует строку Go в формат L2 Unicode (UTF-16LE + null terminator)
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

// l2c1go/internal/gameserver/packets.go
func PackCharSelectionInfo(login string, chars []db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x1F) // opcode CharacterSelectionInfo

	if len(chars) == 0 {
		binary.Write(buf, binary.LittleEndian, uint32(0))
		return buf.Bytes()
	}

	binary.Write(buf, binary.LittleEndian, uint32(len(chars)))

	for _, char := range chars {
		// Каждый персонаж начинается заново с чистой структурой
		buf.Write(encodeUTF16(char.Name))                    // Name
		binary.Write(buf, binary.LittleEndian, uint32(char.ObjectID))
		buf.Write(encodeUTF16(login))                        // Account name
		binary.Write(buf, binary.LittleEndian, uint32(0x55555555)) // Session ID
		binary.Write(buf, binary.LittleEndian, uint32(0))    // Clan ID
		binary.Write(buf, binary.LittleEndian, uint32(0))    // Unknown

		binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
		binary.Write(buf, binary.LittleEndian, uint32(char.Race))
		binary.Write(buf, binary.LittleEndian, uint32(char.Class))
		binary.Write(buf, binary.LittleEndian, uint32(1))    // Active

		// Координаты
		binary.Write(buf, binary.LittleEndian, int32(-70880))
		binary.Write(buf, binary.LittleEndian, int32(257360))
		binary.Write(buf, binary.LittleEndian, int32(-3080))

		// HP / MP
		binary.Write(buf, binary.LittleEndian, float64(100.0))
		binary.Write(buf, binary.LittleEndian, float64(50.0))

		// SP, EXP, Level, Karma
		binary.Write(buf, binary.LittleEndian, uint32(0))
		binary.Write(buf, binary.LittleEndian, uint32(0))
		binary.Write(buf, binary.LittleEndian, uint32(char.Level))
		binary.Write(buf, binary.LittleEndian, uint32(0))

		// 9 неизвестных полей
		for i := 0; i < 9; i++ {
			binary.Write(buf, binary.LittleEndian, uint32(0))
		}

		// Инвентарь (36 полей)
		for i := 0; i < 36; i++ {
			binary.Write(buf, binary.LittleEndian, uint32(0))
		}

		// Внешность
		binary.Write(buf, binary.LittleEndian, uint32(0)) // hairStyle
		binary.Write(buf, binary.LittleEndian, uint32(0)) // hairColor
		binary.Write(buf, binary.LittleEndian, uint32(0)) // face

		// Max HP / Max MP
		binary.Write(buf, binary.LittleEndian, float64(100.0))
		binary.Write(buf, binary.LittleEndian, float64(50.0))

		binary.Write(buf, binary.LittleEndian, uint32(0)) // delete time
	}

	return buf.Bytes()
}

func PackCharSelected(charName string) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x21)
	buf.Write(encodeUTF16(charName))
	binary.Write(buf, binary.LittleEndian, uint32(10001))
	buf.Write(encodeUTF16("Title"))
	binary.Write(buf, binary.LittleEndian, uint32(0x55555555))
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint32(1))

	binary.Write(buf, binary.LittleEndian, int32(-70880))
	binary.Write(buf, binary.LittleEndian, int32(257360))
	binary.Write(buf, binary.LittleEndian, int32(-3080))

	binary.Write(buf, binary.LittleEndian, float64(100.0))
	binary.Write(buf, binary.LittleEndian, float64(50.0))

	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint32(1))
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint32(0))

	binary.Write(buf, binary.LittleEndian, uint32(21))
	binary.Write(buf, binary.LittleEndian, uint32(40))
	binary.Write(buf, binary.LittleEndian, uint32(43))
	binary.Write(buf, binary.LittleEndian, uint32(25))
	binary.Write(buf, binary.LittleEndian, uint32(30))
	binary.Write(buf, binary.LittleEndian, uint32(11))

	for i := 0; i < 30; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(0))
	}

	binary.Write(buf, binary.LittleEndian, uint32(0))
	return buf.Bytes()
}

func PackEnterWorld() []byte {
	return []byte{0x03}
}

func PackQuestList() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x80)
	binary.Write(buf, binary.LittleEndian, uint16(0))
	return buf.Bytes()
}

func PackSkillList() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x58)
	binary.Write(buf, binary.LittleEndian, uint32(0))
	return buf.Bytes()
}