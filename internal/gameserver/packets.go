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

// 0x1F - CharacterSelectionInfo
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
		binary.Write(buf, binary.LittleEndian, uint32(0x55555555)) // SessionID
		binary.Write(buf, binary.LittleEndian, uint32(0))          // ClanID
		binary.Write(buf, binary.LittleEndian, uint32(0))          // Static

		binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
		binary.Write(buf, binary.LittleEndian, uint32(char.Race))
		binary.Write(buf, binary.LittleEndian, uint32(char.Class))
		binary.Write(buf, binary.LittleEndian, uint32(1)) // Active

		binary.Write(buf, binary.LittleEndian, int32(-70880)) // X
		binary.Write(buf, binary.LittleEndian, int32(257360)) // Y
		binary.Write(buf, binary.LittleEndian, int32(-3080))  // Z

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

// 0x15 - CharacterSelected (CharSelected)
func PackCharSelected(char *db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x15) // По твоему описанию тип 0x15

	buf.Write(encodeUTF16(char.Name))
	binary.Write(buf, binary.LittleEndian, uint32(char.ObjectID))
	buf.Write(encodeUTF16("Title"))
	binary.Write(buf, binary.LittleEndian, uint32(0x55555555)) // Session ID
	binary.Write(buf, binary.LittleEndian, uint32(0))          // Clan ID
	binary.Write(buf, binary.LittleEndian, uint32(0))          // Static 0

	binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
	binary.Write(buf, binary.LittleEndian, uint32(char.Race))
	binary.Write(buf, binary.LittleEndian, uint32(char.Class))
	binary.Write(buf, binary.LittleEndian, uint32(1)) // Active

	binary.Write(buf, binary.LittleEndian, int32(-70880)) // X
	binary.Write(buf, binary.LittleEndian, int32(257360)) // Y
	binary.Write(buf, binary.LittleEndian, int32(-3080))  // Z

	binary.Write(buf, binary.LittleEndian, float64(100.0)) // HP (Double)
	binary.Write(buf, binary.LittleEndian, float64(50.0))  // MP (Double)

	binary.Write(buf, binary.LittleEndian, uint32(0))           // SP
	binary.Write(buf, binary.LittleEndian, uint32(0))           // EXP
	binary.Write(buf, binary.LittleEndian, uint32(char.Level))
	binary.Write(buf, binary.LittleEndian, uint32(0))           // Karma
	binary.Write(buf, binary.LittleEndian, uint32(0))           // Static 0

	// Характеристики: STR, DEX, CON, INT, WIT, MEN
	binary.Write(buf, binary.LittleEndian, uint32(40))
	binary.Write(buf, binary.LittleEndian, uint32(30))
	binary.Write(buf, binary.LittleEndian, uint32(43))
	binary.Write(buf, binary.LittleEndian, uint32(21))
	binary.Write(buf, binary.LittleEndian, uint32(11))
	binary.Write(buf, binary.LittleEndian, uint32(25))

	// "Муть" из описания (заполнение нулями)
	for i := 0; i < 32; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(0))
	}

	binary.Write(buf, binary.LittleEndian, uint16(1600)) // in-game time
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Неизвестно

	for i := 0; i < 4; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(0))
	}

	return buf.Bytes()
}

// 0x04 - UserInfo
func PackUserInfo(char *db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x04)

	binary.Write(buf, binary.LittleEndian, int32(-70880)) // X
	binary.Write(buf, binary.LittleEndian, int32(257360)) // Y
	binary.Write(buf, binary.LittleEndian, int32(-3080))  // Z
	binary.Write(buf, binary.LittleEndian, uint32(0))     // Heading

	binary.Write(buf, binary.LittleEndian, uint32(char.ObjectID))
	buf.Write(encodeUTF16(char.Name))

	binary.Write(buf, binary.LittleEndian, uint32(char.Race))
	binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
	binary.Write(buf, binary.LittleEndian, uint32(char.Class))

	binary.Write(buf, binary.LittleEndian, uint32(char.Level))
	binary.Write(buf, binary.LittleEndian, uint32(0))  // EXP
	binary.Write(buf, binary.LittleEndian, uint32(40)) // STR
	binary.Write(buf, binary.LittleEndian, uint32(30)) // DEX
	binary.Write(buf, binary.LittleEndian, uint32(43)) // CON
	binary.Write(buf, binary.LittleEndian, uint32(21)) // INT
	binary.Write(buf, binary.LittleEndian, uint32(11)) // WIT
	binary.Write(buf, binary.LittleEndian, uint32(25)) // MEN

	binary.Write(buf, binary.LittleEndian, uint32(0))     // SP
	binary.Write(buf, binary.LittleEndian, uint32(0))     // Cur Load
	binary.Write(buf, binary.LittleEndian, uint32(10000)) // Max Load
	binary.Write(buf, binary.LittleEndian, uint32(20))    // Unknown C1

	for i := 0; i < 15; i++ { binary.Write(buf, binary.LittleEndian, uint32(0)) } // Paperdoll IDs
	for i := 0; i < 15; i++ { binary.Write(buf, binary.LittleEndian, uint32(0)) } // Paperdoll ObjIDs

	binary.Write(buf, binary.LittleEndian, uint32(4))   // P.Atk
	binary.Write(buf, binary.LittleEndian, uint32(300)) // Atk.Spd
	binary.Write(buf, binary.LittleEndian, uint32(70))  // P.Def
	binary.Write(buf, binary.LittleEndian, uint32(0))   // Evasion
	binary.Write(buf, binary.LittleEndian, uint32(0))   // Accuracy
	binary.Write(buf, binary.LittleEndian, uint32(40))  // Critical
	binary.Write(buf, binary.LittleEndian, uint32(3))   // M.Atk
	binary.Write(buf, binary.LittleEndian, uint32(200)) // Cast.Spd

	binary.Write(buf, binary.LittleEndian, uint32(115)) // Speed
	binary.Write(buf, binary.LittleEndian, uint32(115)) // RunSpeed
	binary.Write(buf, binary.LittleEndian, uint32(50))  // WalkSpeed
	binary.Write(buf, binary.LittleEndian, uint32(115)) // SwimRun
	binary.Write(buf, binary.LittleEndian, uint32(50))  // SwimWalk

	binary.Write(buf, binary.LittleEndian, float64(1.0)) // Move Mult
	binary.Write(buf, binary.LittleEndian, float64(1.0)) // Atk Mult

	binary.Write(buf, binary.LittleEndian, float64(8.0))  // Radius
	binary.Write(buf, binary.LittleEndian, float64(24.0)) // Height

	binary.Write(buf, binary.LittleEndian, uint32(0)) // Hair Style
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Hair Color
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Face
	binary.Write(buf, binary.LittleEndian, uint32(0)) // IsGM

	buf.Write(encodeUTF16("Title"))
	binary.Write(buf, binary.LittleEndian, uint32(0)) // ClanID
	binary.Write(buf, binary.LittleEndian, uint32(0)) // AllyID

	return buf.Bytes()
}

// 0xF8 - SSQInfo (SignsSky)
func PackSSQInfo() []byte {
	return []byte{0xF8, 0x01, 0x01} // Dusk победил
}

func PackQuestList() []byte {
	return []byte{0x80, 0x00, 0x00}
}

func PackSkillList() []byte {
	return []byte{0x58, 0x00, 0x00, 0x00, 0x00}
}
