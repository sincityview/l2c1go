// l2c1go/internal/gameserver
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

func PackCharSelectionInfo(login string, chars []db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x1F)

	if len(chars) == 0 {
		binary.Write(buf, binary.LittleEndian, uint32(0))
		return buf.Bytes()
	}

	binary.Write(buf, binary.LittleEndian, uint32(len(chars)))

	for _, char := range chars {
		buf.Write(encodeUTF16(char.Name))
		binary.Write(buf, binary.LittleEndian, uint32(char.ObjectID))
		buf.Write(encodeUTF16(login))
		binary.Write(buf, binary.LittleEndian, uint32(0x55555555)) // sessionId
		binary.Write(buf, binary.LittleEndian, uint32(0))          // clanId
		binary.Write(buf, binary.LittleEndian, uint32(0))          // static 0

		binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
		binary.Write(buf, binary.LittleEndian, uint32(char.Race))
		binary.Write(buf, binary.LittleEndian, uint32(char.Class))
		binary.Write(buf, binary.LittleEndian, uint32(1)) // active

		binary.Write(buf, binary.LittleEndian, int32(-70880)) // x
		binary.Write(buf, binary.LittleEndian, int32(257360)) // y
		binary.Write(buf, binary.LittleEndian, int32(-3080))  // z

		binary.Write(buf, binary.LittleEndian, float64(100.0)) // hp (double)
		binary.Write(buf, binary.LittleEndian, float64(50.0))  // mp (double)

		binary.Write(buf, binary.LittleEndian, uint32(0)) // sp
		binary.Write(buf, binary.LittleEndian, uint32(0)) // exp
		binary.Write(buf, binary.LittleEndian, uint32(char.Level))
		binary.Write(buf, binary.LittleEndian, uint32(0)) // karma

		// Блок заглушек из JS (9 пустых D)
		for i := 0; i < 9; i++ {
			binary.Write(buf, binary.LittleEndian, uint32(0))
		}

		// Блок инвентаря из JS (два раза по 18 пустых D)
		// Первый раз - ObjectIDs, второй раз - ItemIDs
		for i := 0; i < 36; i++ {
			binary.Write(buf, binary.LittleEndian, uint32(0))
		}

		binary.Write(buf, binary.LittleEndian, uint32(0)) // hairStyle
		binary.Write(buf, binary.LittleEndian, uint32(0)) // hairColor
		binary.Write(buf, binary.LittleEndian, uint32(0)) // face

		binary.Write(buf, binary.LittleEndian, float64(100.0)) // maxHp
		binary.Write(buf, binary.LittleEndian, float64(50.0))  // maxMp

		binary.Write(buf, binary.LittleEndian, uint32(0)) // days left before delete (0 - стоит)
	}

	return buf.Bytes()
}



// Пакет 0x17: Character Templates
func PackCharTemplates(templatesData []byte) []byte {
    // Тут будет логика сборки 0x17, если захочешь вынести и её
    return templatesData
}

func PackCharSelected(charName string) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x21) // ID из твоего JS: 0x21

	buf.Write(encodeUTF16(charName))
	binary.Write(buf, binary.LittleEndian, uint32(10001)) // objectId
	buf.Write(encodeUTF16("Title"))                      // title
	binary.Write(buf, binary.LittleEndian, uint32(0x55555555))
	binary.Write(buf, binary.LittleEndian, uint32(0)) // clanId
	binary.Write(buf, binary.LittleEndian, uint32(0)) // static
	binary.Write(buf, binary.LittleEndian, uint32(0)) // gender
	binary.Write(buf, binary.LittleEndian, uint32(0)) // raceId
	binary.Write(buf, binary.LittleEndian, uint32(0)) // classId
	binary.Write(buf, binary.LittleEndian, uint32(1)) // active

	binary.Write(buf, binary.LittleEndian, int32(-70880)) // x
	binary.Write(buf, binary.LittleEndian, int32(257360)) // y
	binary.Write(buf, binary.LittleEndian, int32(-3080))  // z

	binary.Write(buf, binary.LittleEndian, float64(100.0)) // hp
	binary.Write(buf, binary.LittleEndian, float64(50.0))  // mp

	binary.Write(buf, binary.LittleEndian, uint32(0)) // sp
	binary.Write(buf, binary.LittleEndian, uint32(0)) // exp
	binary.Write(buf, binary.LittleEndian, uint32(1)) // level
	binary.Write(buf, binary.LittleEndian, uint32(0)) // static 0
	binary.Write(buf, binary.LittleEndian, uint32(0)) // static 0

	// Статы: INT, STR, CON, MEN, DEX, WIT
	binary.Write(buf, binary.LittleEndian, uint32(21)) // int
	binary.Write(buf, binary.LittleEndian, uint32(40)) // str
	binary.Write(buf, binary.LittleEndian, uint32(43)) // con
	binary.Write(buf, binary.LittleEndian, uint32(25)) // men
	binary.Write(buf, binary.LittleEndian, uint32(30)) // dex
	binary.Write(buf, binary.LittleEndian, uint32(11)) // wit

	// Цикл из 30 пустых D (как в JS)
	for i := 0; i < 30; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(0))
	}

	binary.Write(buf, binary.LittleEndian, uint32(0)) // in-game time

	return buf.Bytes()
}


// Пакет 0x03: EnterWorld
func PackEnterWorld() []byte {
    return []byte{0x03} // В C1 это часто просто ID пакета
}

// Пакет 0x80: QuestList (в C1 часто пустой список)
func PackQuestList() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x80) // ID
	binary.Write(buf, binary.LittleEndian, uint16(0)) // Количество квестов
	return buf.Bytes()
}

// Пакет 0x58: SkillList
func PackSkillList() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x58) // ID
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Количество скиллов
	return buf.Bytes()
}
