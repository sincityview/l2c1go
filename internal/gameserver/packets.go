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
		binary.Write(buf, binary.LittleEndian, uint32(0x55555555))
		binary.Write(buf, binary.LittleEndian, uint32(0)) 
		binary.Write(buf, binary.LittleEndian, uint32(0))

		binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
		binary.Write(buf, binary.LittleEndian, uint32(char.Race))
		binary.Write(buf, binary.LittleEndian, uint32(char.ClassID)) // Исправлено на ClassID
		binary.Write(buf, binary.LittleEndian, uint32(1)) 

		binary.Write(buf, binary.LittleEndian, int32(char.X))
		binary.Write(buf, binary.LittleEndian, int32(char.Y))
		binary.Write(buf, binary.LittleEndian, int32(char.Z))

		binary.Write(buf, binary.LittleEndian, float64(char.CurHp))
		binary.Write(buf, binary.LittleEndian, float64(char.CurMp))

		binary.Write(buf, binary.LittleEndian, uint32(char.Sp))
		binary.Write(buf, binary.LittleEndian, uint32(char.Exp))
		binary.Write(buf, binary.LittleEndian, uint32(char.Level))
		binary.Write(buf, binary.LittleEndian, uint32(char.Karma))

		for i := 0; i < 9; i++ { binary.Write(buf, binary.LittleEndian, uint32(0)) }
		for i := 0; i < 36; i++ { binary.Write(buf, binary.LittleEndian, uint32(0)) }

		binary.Write(buf, binary.LittleEndian, uint32(char.HairStyle))
		binary.Write(buf, binary.LittleEndian, uint32(char.HairColor))
		binary.Write(buf, binary.LittleEndian, uint32(char.Face))
		binary.Write(buf, binary.LittleEndian, float64(char.MaxHp))
		binary.Write(buf, binary.LittleEndian, float64(char.MaxMp))
		binary.Write(buf, binary.LittleEndian, uint32(0)) 
	}
	return buf.Bytes()
}

// 0x21 - CharacterSelected
func PackCharSelected(char *db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x21)

	buf.Write(encodeUTF16(char.Name))
	binary.Write(buf, binary.LittleEndian, uint32(char.ObjectID))
	buf.Write(encodeUTF16(char.Title)) // Теперь берем титул из базы
	binary.Write(buf, binary.LittleEndian, uint32(0x55555555))
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint32(0))
	binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
	binary.Write(buf, binary.LittleEndian, uint32(char.Race))
	binary.Write(buf, binary.LittleEndian, uint32(char.ClassID)) // Исправлено на ClassID
	binary.Write(buf, binary.LittleEndian, uint32(1)) 

	binary.Write(buf, binary.LittleEndian, int32(char.X))
	binary.Write(buf, binary.LittleEndian, int32(char.Y))
	binary.Write(buf, binary.LittleEndian, int32(char.Z))

	binary.Write(buf, binary.LittleEndian, float64(char.CurHp))
	binary.Write(buf, binary.LittleEndian, float64(char.CurMp))

	binary.Write(buf, binary.LittleEndian, uint32(char.Sp))
	binary.Write(buf, binary.LittleEndian, uint32(char.Exp))
	binary.Write(buf, binary.LittleEndian, uint32(char.Level))
	binary.Write(buf, binary.LittleEndian, uint32(char.Karma))
	binary.Write(buf, binary.LittleEndian, uint32(0)) 

	binary.Write(buf, binary.LittleEndian, uint32(21)) // INT
	binary.Write(buf, binary.LittleEndian, uint32(40)) // STR
	binary.Write(buf, binary.LittleEndian, uint32(43)) // CON
	binary.Write(buf, binary.LittleEndian, uint32(25)) // MEN
	binary.Write(buf, binary.LittleEndian, uint32(30)) // DEX
	binary.Write(buf, binary.LittleEndian, uint32(11)) // WIT

	for i := 0; i < 30; i++ { binary.Write(buf, binary.LittleEndian, uint32(0)) }
	binary.Write(buf, binary.LittleEndian, uint32(0)) 
	return buf.Bytes()
}

func PackUserInfo(char *db.CharData, paperdollObj [15]int32, paperdollItem [15]int32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x04)

	binary.Write(buf, binary.LittleEndian, int32(char.X))
	binary.Write(buf, binary.LittleEndian, int32(char.Y))
	binary.Write(buf, binary.LittleEndian, int32(char.Z))
	binary.Write(buf, binary.LittleEndian, uint32(0)) 
	binary.Write(buf, binary.LittleEndian, uint32(char.ObjectID))
	buf.Write(encodeUTF16(char.Name))
	binary.Write(buf, binary.LittleEndian, uint32(char.Race))
	binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
	binary.Write(buf, binary.LittleEndian, uint32(char.ClassID))
	binary.Write(buf, binary.LittleEndian, uint32(char.Level))
	binary.Write(buf, binary.LittleEndian, uint32(char.Exp))

	binary.Write(buf, binary.LittleEndian, uint32(40)) // STR заглушка
	binary.Write(buf, binary.LittleEndian, uint32(30)) 
	binary.Write(buf, binary.LittleEndian, uint32(43)) 
	binary.Write(buf, binary.LittleEndian, uint32(21)) 
	binary.Write(buf, binary.LittleEndian, uint32(11)) 
	binary.Write(buf, binary.LittleEndian, uint32(25)) 

	binary.Write(buf, binary.LittleEndian, uint32(char.MaxHp))
	binary.Write(buf, binary.LittleEndian, uint32(char.CurHp))
	binary.Write(buf, binary.LittleEndian, uint32(char.MaxMp))
	binary.Write(buf, binary.LittleEndian, uint32(char.CurMp))

	binary.Write(buf, binary.LittleEndian, uint32(char.Sp))
	binary.Write(buf, binary.LittleEndian, uint32(100)) 
	binary.Write(buf, binary.LittleEndian, uint32(1000))
	binary.Write(buf, binary.LittleEndian, uint32(0x28)) 

	// ВАЖНО: Выводим ObjectID предметов
	for i := 0; i < 15; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(paperdollObj[i]))
	}
	// ВАЖНО: Выводим ItemID предметов
	for i := 0; i < 15; i++ {
		binary.Write(buf, binary.LittleEndian, uint32(paperdollItem[i]))
	}

	binary.Write(buf, binary.LittleEndian, uint32(4))   // pAtk
	binary.Write(buf, binary.LittleEndian, uint32(300)) // attackSpeed
	binary.Write(buf, binary.LittleEndian, uint32(70))  // pDef
	binary.Write(buf, binary.LittleEndian, uint32(0))   
	binary.Write(buf, binary.LittleEndian, uint32(0))   
	binary.Write(buf, binary.LittleEndian, uint32(40))  
	binary.Write(buf, binary.LittleEndian, uint32(3))   
	binary.Write(buf, binary.LittleEndian, uint32(200))
	binary.Write(buf, binary.LittleEndian, uint32(300))
	binary.Write(buf, binary.LittleEndian, uint32(70)) 
	binary.Write(buf, binary.LittleEndian, uint32(0)) 
	binary.Write(buf, binary.LittleEndian, uint32(char.Karma))

	for i := 0; i < 8; i++ { binary.Write(buf, binary.LittleEndian, uint32(115)) }
	binary.Write(buf, binary.LittleEndian, float64(1.0)) 
	binary.Write(buf, binary.LittleEndian, float64(1.0))
	binary.Write(buf, binary.LittleEndian, float64(8.0)) 
	binary.Write(buf, binary.LittleEndian, float64(24.0))

	binary.Write(buf, binary.LittleEndian, uint32(char.HairStyle))
	binary.Write(buf, binary.LittleEndian, uint32(char.HairColor))
	binary.Write(buf, binary.LittleEndian, uint32(char.Face))
	binary.Write(buf, binary.LittleEndian, uint32(0)) 
	buf.Write(encodeUTF16(char.Title))
	binary.Write(buf, binary.LittleEndian, uint32(0)) 
	binary.Write(buf, binary.LittleEndian, uint32(0)) 
	binary.Write(buf, binary.LittleEndian, uint32(0)) 
	binary.Write(buf, binary.LittleEndian, uint32(0)) 
	binary.Write(buf, binary.LittleEndian, uint32(0)) 
	buf.WriteByte(0x00) 
	buf.WriteByte(0x00) 
	buf.WriteByte(0x00) 
	binary.Write(buf, binary.LittleEndian, uint32(0)) 
	binary.Write(buf, binary.LittleEndian, uint32(0)) 
	binary.Write(buf, binary.LittleEndian, uint16(0)) 
	buf.WriteByte(0x00) 

	return buf.Bytes()
}

func PackItemList(items []db.ItemData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x27) // ПРАВИЛЬНЫЙ ОПКОД ДЛЯ C1

	binary.Write(buf, binary.LittleEndian, uint16(1)) // showWindow = true
	binary.Write(buf, binary.LittleEndian, uint16(len(items)))

	for _, it := range items {
		binary.Write(buf, binary.LittleEndian, uint16(0)) // getType1()
		binary.Write(buf, binary.LittleEndian, uint32(it.ObjectID))
		binary.Write(buf, binary.LittleEndian, uint32(it.ItemID))
		binary.Write(buf, binary.LittleEndian, uint32(it.Count))
		binary.Write(buf, binary.LittleEndian, uint16(0)) // getType2()
		
		binary.Write(buf, binary.LittleEndian, uint16(0xFF)) // ТА САМАЯ ЗАГЛУШКА (writeH(0xff))

		isEquipped := uint16(0)
		if it.Loc == "PAPERDOLL" {
			isEquipped = 1
		}
		binary.Write(buf, binary.LittleEndian, isEquipped) 
		
		binary.Write(buf, binary.LittleEndian, uint32(0)) // getBodyPart()
		binary.Write(buf, binary.LittleEndian, uint16(it.EnchantLevel))
		binary.Write(buf, binary.LittleEndian, uint16(0)) // Еще одна заглушка в конце
	}
	return buf.Bytes()
}

func PackQuestList() []byte { return []byte{0x80, 0x00, 0x00} }
func PackSkillList() []byte { return []byte{0x58, 0x00, 0x00, 0x00, 0x00} }
func PackSSQInfo() []byte    { return []byte{0xF8, 0x01, 0x01} }

func PackSystemMessage(msgID uint32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x7A)
	binary.Write(buf, binary.LittleEndian, msgID)
	binary.Write(buf, binary.LittleEndian, uint32(0))
	return buf.Bytes()
}

func PackCharMoveToLocation(objID int32, targetX, targetY, targetZ int32, curX, curY, curZ int32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x01)
	binary.Write(buf, binary.LittleEndian, objID)
	binary.Write(buf, binary.LittleEndian, targetX)
	binary.Write(buf, binary.LittleEndian, targetY)
	binary.Write(buf, binary.LittleEndian, targetZ)
	binary.Write(buf, binary.LittleEndian, curX)
	binary.Write(buf, binary.LittleEndian, curY)
	binary.Write(buf, binary.LittleEndian, curZ)
	return buf.Bytes()
}

// 0x5D - Say2 (Серверный пакет чата)
func PackSay2(objID int32, chatType uint32, charName string, text string) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x5D) // Твой проверенный опкод для C1
	binary.Write(buf, binary.LittleEndian, objID)
	binary.Write(buf, binary.LittleEndian, chatType)
	buf.Write(encodeUTF16(charName))
	buf.Write(encodeUTF16(text))
	return buf.Bytes()
}

// 0x3D - SocialAction (Серверный пакет анимации)
func PackSocialAction(objID int32, actionID uint32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x3D)
	binary.Write(buf, binary.LittleEndian, objID)
	binary.Write(buf, binary.LittleEndian, actionID)
	return buf.Bytes()
}

// 0xBD - ShowRadar (Стрелка на радаре)
func PackShowRadar(x, y, z int32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0xBD)
	binary.Write(buf, binary.LittleEndian, x)
	binary.Write(buf, binary.LittleEndian, y)
	binary.Write(buf, binary.LittleEndian, z)
	return buf.Bytes()
}

// 0xBF - MyTargetSelected
func PackMyTargetSelected(objID int32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0xBF)
	binary.Write(buf, binary.LittleEndian, objID)
	binary.Write(buf, binary.LittleEndian, uint16(0)) // Разница в уровнях (0 - белый цвет)
	return buf.Bytes()
}

// 0x3A - TargetUnselected (Отмена таргета)
func PackTargetUnselected(char *db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x3A)
	binary.Write(buf, binary.LittleEndian, char.ObjectID)
	binary.Write(buf, binary.LittleEndian, int32(char.X))
	binary.Write(buf, binary.LittleEndian, int32(char.Y))
	binary.Write(buf, binary.LittleEndian, int32(char.Z))
	binary.Write(buf, binary.LittleEndian, uint32(0)) 
	return buf.Bytes()
}

// 0xB6 - ShowMiniMap (Карта)
func PackShowMiniMap(mapID int32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0xB6)
	binary.Write(buf, binary.LittleEndian, mapID)
	return buf.Bytes()
}

// 0x86 - ShowBoard (Alt+B)
func PackShowBoard(html string) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x86)
	for i := 0; i < 6; i++ { buf.Write(encodeUTF16("")) }
	buf.Write(encodeUTF16(html))
	return buf.Bytes()
}

