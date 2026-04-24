package gameserver

import (
	"bytes"
	"encoding/binary"
	"l2c1go/internal/db"
)

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
		binary.Write(buf, binary.LittleEndian, uint32(0x55555555))
		binary.Write(buf, binary.LittleEndian, uint32(0)) // ClanID
		binary.Write(buf, binary.LittleEndian, uint32(0))

		binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
		binary.Write(buf, binary.LittleEndian, uint32(char.Race))
		binary.Write(buf, binary.LittleEndian, uint32(char.Class))
		binary.Write(buf, binary.LittleEndian, uint32(1)) // Active

		binary.Write(buf, binary.LittleEndian, int32(char.X))
		binary.Write(buf, binary.LittleEndian, int32(char.Y))
		binary.Write(buf, binary.LittleEndian, int32(char.Z))

		binary.Write(buf, binary.LittleEndian, float64(100.0)) // HP
		binary.Write(buf, binary.LittleEndian, float64(50.0))  // MP

		binary.Write(buf, binary.LittleEndian, uint32(0)) // SP
		binary.Write(buf, binary.LittleEndian, uint32(0)) // EXP
		binary.Write(buf, binary.LittleEndian, uint32(char.Level))
		binary.Write(buf, binary.LittleEndian, uint32(0)) // Karma

		for i := 0; i < 9; i++ { binary.Write(buf, binary.LittleEndian, uint32(0)) }
		for i := 0; i < 36; i++ { binary.Write(buf, binary.LittleEndian, uint32(0)) }

		binary.Write(buf, binary.LittleEndian, uint32(0))      // hairStyle
		binary.Write(buf, binary.LittleEndian, uint32(0))      // hairColor
		binary.Write(buf, binary.LittleEndian, uint32(0))      // face
		binary.Write(buf, binary.LittleEndian, float64(100.0)) // maxHp
		binary.Write(buf, binary.LittleEndian, float64(50.0))  // maxMp
		binary.Write(buf, binary.LittleEndian, uint32(0))      // delete time
	}
	return buf.Bytes()
}



// 0x21 - CharacterSelected (из JS-проекта)
func PackCharSelected(char *db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x21)

	buf.Write(encodeUTF16(char.Name))
	binary.Write(buf, binary.LittleEndian, uint32(char.ObjectID))
	buf.Write(encodeUTF16("Title"))
	binary.Write(buf, binary.LittleEndian, uint32(0x55555555))
	binary.Write(buf, binary.LittleEndian, uint32(0)) // ClanId
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Static 0
	binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
	binary.Write(buf, binary.LittleEndian, uint32(char.Race))
	binary.Write(buf, binary.LittleEndian, uint32(char.Class))
	binary.Write(buf, binary.LittleEndian, uint32(1)) // Active

	binary.Write(buf, binary.LittleEndian, int32(char.X))
	binary.Write(buf, binary.LittleEndian, int32(char.Y))
	binary.Write(buf, binary.LittleEndian, int32(char.Z))

	binary.Write(buf, binary.LittleEndian, float64(100.0)) // HP (Float)
	binary.Write(buf, binary.LittleEndian, float64(50.0))  // MP (Float)

	binary.Write(buf, binary.LittleEndian, uint32(0)) // SP
	binary.Write(buf, binary.LittleEndian, uint32(0)) // EXP
	binary.Write(buf, binary.LittleEndian, uint32(char.Level))
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Unknown/Karma
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Unknown

	// Порядок статов из JS: INT, STR, CON, MEN, DEX, WIT
	binary.Write(buf, binary.LittleEndian, uint32(21)) // INT
	binary.Write(buf, binary.LittleEndian, uint32(40)) // STR
	binary.Write(buf, binary.LittleEndian, uint32(43)) // CON
	binary.Write(buf, binary.LittleEndian, uint32(25)) // MEN
	binary.Write(buf, binary.LittleEndian, uint32(30)) // DEX
	binary.Write(buf, binary.LittleEndian, uint32(11)) // WIT

	for i := 0; i < 30; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(0))
	}
	binary.Write(buf, binary.LittleEndian, uint32(0)) // in-game time
	return buf.Bytes()
}

// 0x04 - UserInfo (Критическая структура для входа)
func PackUserInfo(char *db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x04)

	binary.Write(buf, binary.LittleEndian, int32(char.X))
	binary.Write(buf, binary.LittleEndian, int32(char.Y))
	binary.Write(buf, binary.LittleEndian, int32(char.Z))
	binary.Write(buf, binary.LittleEndian, uint32(0)) // heading
	binary.Write(buf, binary.LittleEndian, uint32(char.ObjectID))
	buf.Write(encodeUTF16(char.Name))
	binary.Write(buf, binary.LittleEndian, uint32(char.Race))
	binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
	binary.Write(buf, binary.LittleEndian, uint32(char.Class))
	binary.Write(buf, binary.LittleEndian, uint32(char.Level))
	binary.Write(buf, binary.LittleEndian, uint32(0)) // exp

	// Статы в UserInfo: STR, DEX, CON, INT, WIT, MEN
	binary.Write(buf, binary.LittleEndian, uint32(40)) // STR
	binary.Write(buf, binary.LittleEndian, uint32(30)) // DEX
	binary.Write(buf, binary.LittleEndian, uint32(43)) // CON
	binary.Write(buf, binary.LittleEndian, uint32(21)) // INT
	binary.Write(buf, binary.LittleEndian, uint32(11)) // WIT
	binary.Write(buf, binary.LittleEndian, uint32(25)) // MEN

	// HP/MP — тут writeD (uint32) по JS-проекту!
	binary.Write(buf, binary.LittleEndian, uint32(100)) // Max HP
	binary.Write(buf, binary.LittleEndian, uint32(100)) // Cur HP
	binary.Write(buf, binary.LittleEndian, uint32(50))  // Max MP
	binary.Write(buf, binary.LittleEndian, uint32(50))  // Cur MP

	binary.Write(buf, binary.LittleEndian, uint32(0))     // sp
	binary.Write(buf, binary.LittleEndian, uint32(100))   // cur load
	binary.Write(buf, binary.LittleEndian, uint32(1000))  // max load
	binary.Write(buf, binary.LittleEndian, uint32(0x28))  // unknown static

	// Paperdoll (2 блока по 15 слотов: ObjectId и ItemId) = 30 * writeD
	for i := 0; i < 30; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(0))
	}

	binary.Write(buf, binary.LittleEndian, uint32(4))   // pAtk
	binary.Write(buf, binary.LittleEndian, uint32(300)) // attackSpeed
	binary.Write(buf, binary.LittleEndian, uint32(70))  // pDef
	binary.Write(buf, binary.LittleEndian, uint32(0))   // evasion
	binary.Write(buf, binary.LittleEndian, uint32(0))   // accuracy
	binary.Write(buf, binary.LittleEndian, uint32(40))  // critical
	binary.Write(buf, binary.LittleEndian, uint32(3))   // mAtk
	binary.Write(buf, binary.LittleEndian, uint32(200)) // mSpd
	binary.Write(buf, binary.LittleEndian, uint32(300)) // pSpd (дубль?)
	binary.Write(buf, binary.LittleEndian, uint32(70))  // mDef
	binary.Write(buf, binary.LittleEndian, uint32(0))   // purple
	binary.Write(buf, binary.LittleEndian, uint32(0))   // karma

	// Скорости (8 штук writeD)
	for i := 0; i < 8; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(115))
	}

	binary.Write(buf, binary.LittleEndian, float64(1.0)) // movementMultiplier
	binary.Write(buf, binary.LittleEndian, float64(1.0)) // attackSpeedMultiplier
	binary.Write(buf, binary.LittleEndian, float64(8.0)) // collisionRadius
	binary.Write(buf, binary.LittleEndian, float64(24.0)) // collisionHeight

	binary.Write(buf, binary.LittleEndian, uint32(0)) // hairStyle
	binary.Write(buf, binary.LittleEndian, uint32(0)) // hairColor
	binary.Write(buf, binary.LittleEndian, uint32(0)) // face
	binary.Write(buf, binary.LittleEndian, uint32(0)) // isGM
	buf.Write(encodeUTF16("Title"))
	binary.Write(buf, binary.LittleEndian, uint32(0)) // clanId
	binary.Write(buf, binary.LittleEndian, uint32(0)) // crestId
	binary.Write(buf, binary.LittleEndian, uint32(0)) // allyId
	binary.Write(buf, binary.LittleEndian, uint32(0)) // allyCrestId
	binary.Write(buf, binary.LittleEndian, uint32(0)) // siege-flags
	buf.WriteByte(0x00) // byte
	buf.WriteByte(0x00) // storeType
	buf.WriteByte(0x00) // canCraft
	binary.Write(buf, binary.LittleEndian, uint32(0)) // pk
	binary.Write(buf, binary.LittleEndian, uint32(0)) // pvp
	binary.Write(buf, binary.LittleEndian, uint16(0)) // cubic count
	buf.WriteByte(0x00) // find party

	return buf.Bytes()
}

func PackItemList() []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x1B)
	binary.Write(buf, binary.LittleEndian, uint16(0))
	return buf.Bytes()
}

func PackQuestList() []byte {
	return []byte{0x80, 0x00, 0x00}
}

func PackSkillList() []byte {
	return []byte{0x58, 0x00, 0x00, 0x00, 0x00}
}

func PackSSQInfo() []byte {
	return []byte{0xF8, 0x01, 0x01}
}

func PackCharMoveToLocation(objID int32, targetX, targetY, targetZ int32, curX, curY, curZ int32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x01) // Opcode MoveToLocation
	binary.Write(buf, binary.LittleEndian, objID)
	binary.Write(buf, binary.LittleEndian, targetX)
	binary.Write(buf, binary.LittleEndian, targetY)
	binary.Write(buf, binary.LittleEndian, targetZ)
	binary.Write(buf, binary.LittleEndian, curX)
	binary.Write(buf, binary.LittleEndian, curY)
	binary.Write(buf, binary.LittleEndian, curZ)
	return buf.Bytes()
}

func PackSay2(objID int32, chatType uint32, charName string, text string) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x5D) // Наш новый опкод
	binary.Write(buf, binary.LittleEndian, objID)
	binary.Write(buf, binary.LittleEndian, chatType)
	buf.Write(encodeUTF16(charName))
	buf.Write(encodeUTF16(text))
	return buf.Bytes()
}

func PackSystemMessage(msgID uint32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x7A)
	binary.Write(buf, binary.LittleEndian, msgID)
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Количество параметров (в С1 это важно)
	return buf.Bytes()
}