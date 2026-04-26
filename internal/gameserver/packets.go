package gameserver

import (
	"bytes"
	"encoding/binary"
	"darkages/internal/db"
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

// 0x1F - CharSelectionInfo (Лобби)
func PackCharSelectionInfo(login string, chars []db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x1F)

	if len(chars) == 0 {
		binary.Write(buf, binary.LittleEndian, uint32(0))
		return buf.Bytes()
	}

	binary.Write(buf, binary.LittleEndian, uint32(len(chars)))

	for _, char := range chars {
		// ПОЛУЧАЕМ ПРЕДМЕТЫ ДЛЯ ОТОБРАЖЕНИЯ В ЛОББИ
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

		// 1. Блок из 9 нулей (Reserved)
		for i := 0; i < 9; i++ {
			binary.Write(buf, binary.LittleEndian, uint32(0))
		}

		// 2. Блок из 15 ObjectID (Кукла)
		for i := 0; i < 15; i++ {
			binary.Write(buf, binary.LittleEndian, uint32(pObj[i]))
		}

		// 3. Блок из 15 ItemID (Кукла)
		for i := 0; i < 15; i++ {
			binary.Write(buf, binary.LittleEndian, uint32(pItem[i]))
		}

		binary.Write(buf, binary.LittleEndian, uint32(char.HairStyle))
		binary.Write(buf, binary.LittleEndian, uint32(char.HairColor))
		binary.Write(buf, binary.LittleEndian, uint32(char.Face))

		binary.Write(buf, binary.LittleEndian, float64(char.MaxHp))
		binary.Write(buf, binary.LittleEndian, float64(char.MaxMp))

		binary.Write(buf, binary.LittleEndian, uint32(0)) // Флаг удаления
	}
	return buf.Bytes()
}

func PackCharSelected(char *db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x21)

	buf.Write(encodeUTF16(char.Name))
	binary.Write(buf, binary.LittleEndian, uint32(char.ObjectID))
	buf.Write(encodeUTF16(char.Title))
	binary.Write(buf, binary.LittleEndian, uint32(0x55555555)) // Session
	binary.Write(buf, binary.LittleEndian, uint32(0))          // Clan
	binary.Write(buf, binary.LittleEndian, uint32(0))          // Placeholder
	
	binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
	binary.Write(buf, binary.LittleEndian, uint32(char.Race))
	binary.Write(buf, binary.LittleEndian, uint32(char.ClassID))
	binary.Write(buf, binary.LittleEndian, uint32(1))          // Active

	binary.Write(buf, binary.LittleEndian, int32(char.X))
	binary.Write(buf, binary.LittleEndian, int32(char.Y))
	binary.Write(buf, binary.LittleEndian, int32(char.Z))

	// В С1 ТУТ СТРОГО FLOAT64 (8 байт каждое)
	binary.Write(buf, binary.LittleEndian, float64(char.CurHp))
	binary.Write(buf, binary.LittleEndian, float64(char.CurMp))

	binary.Write(buf, binary.LittleEndian, uint32(char.Sp))
	binary.Write(buf, binary.LittleEndian, uint32(char.Exp))
	binary.Write(buf, binary.LittleEndian, uint32(char.Level))
	binary.Write(buf, binary.LittleEndian, uint32(char.Karma))
	binary.Write(buf, binary.LittleEndian, uint32(0)) // PK
	
	// В С1 после PK идет 16 байт заглушек (4 по writeD(0))
	for i := 0; i < 4; i++ { binary.Write(buf, binary.LittleEndian, uint32(0)) }

	return buf.Bytes()
}

func PackUserInfo(char *db.CharData, paperdollObj [15]int32, paperdollItem [15]int32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x04)

	binary.Write(buf, binary.LittleEndian, int32(char.X))
	binary.Write(buf, binary.LittleEndian, int32(char.Y))
	binary.Write(buf, binary.LittleEndian, int32(char.Z))
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Heading
	binary.Write(buf, binary.LittleEndian, uint32(char.ObjectID))
	buf.Write(encodeUTF16(char.Name))
	binary.Write(buf, binary.LittleEndian, uint32(char.Race))
	binary.Write(buf, binary.LittleEndian, uint32(char.Sex))
	binary.Write(buf, binary.LittleEndian, uint32(char.ClassID))
	binary.Write(buf, binary.LittleEndian, uint32(char.Level))
	binary.Write(buf, binary.LittleEndian, uint32(char.Exp))

	// Базовые статы
	binary.Write(buf, binary.LittleEndian, uint32(40)) // STR
	binary.Write(buf, binary.LittleEndian, uint32(30)) // DEX
	binary.Write(buf, binary.LittleEndian, uint32(43)) // CON
	binary.Write(buf, binary.LittleEndian, uint32(21)) // INT
	binary.Write(buf, binary.LittleEndian, uint32(11)) // WIT
	binary.Write(buf, binary.LittleEndian, uint32(25)) // MEN

	binary.Write(buf, binary.LittleEndian, uint32(char.MaxHp))
	binary.Write(buf, binary.LittleEndian, uint32(char.CurHp))
	binary.Write(buf, binary.LittleEndian, uint32(char.MaxMp))
	binary.Write(buf, binary.LittleEndian, uint32(char.CurMp))

	binary.Write(buf, binary.LittleEndian, uint32(char.Sp))
	binary.Write(buf, binary.LittleEndian, uint32(0))    // Current Load
	binary.Write(buf, binary.LittleEndian, uint32(1000)) // Max Load
	binary.Write(buf, binary.LittleEndian, uint32(0x14)) // Режим (20 в С1)

	// Бумажная кукла (15 ObjectID)
	for i := 0; i < 15; i++ {
		binary.Write(buf, binary.LittleEndian, int32(paperdollObj[i]))
	}
	// Бумажная кукла (15 ItemID)
	for i := 0; i < 15; i++ {
		binary.Write(buf, binary.LittleEndian, int32(paperdollItem[i]))
	}

	// Параметры боя
	binary.Write(buf, binary.LittleEndian, uint32(4))   // PAtk
	binary.Write(buf, binary.LittleEndian, uint32(300)) // AtkSpd
	binary.Write(buf, binary.LittleEndian, uint32(70))  // PDef
	binary.Write(buf, binary.LittleEndian, uint32(0))   // Evasion
	binary.Write(buf, binary.LittleEndian, uint32(0))   // Accuracy
	binary.Write(buf, binary.LittleEndian, uint32(40))  // Critical
	binary.Write(buf, binary.LittleEndian, uint32(3))   // MAtk
	binary.Write(buf, binary.LittleEndian, uint32(200)) // CastSpd
	binary.Write(buf, binary.LittleEndian, uint32(300)) // Speed
	binary.Write(buf, binary.LittleEndian, uint32(70))  // MDef
	binary.Write(buf, binary.LittleEndian, uint32(0))   // Пурпурный ник
	binary.Write(buf, binary.LittleEndian, uint32(char.Karma))

	// Скорости (Run, Walk, Swim, SwimWalk)
	binary.Write(buf, binary.LittleEndian, uint32(115)) // Run
	binary.Write(buf, binary.LittleEndian, uint32(115)) // Walk
	binary.Write(buf, binary.LittleEndian, uint32(115)) // Swim
	binary.Write(buf, binary.LittleEndian, uint32(115)) // Swim Walk

	// Заглушки для старых хроник (4 по 4 байта)
	for i := 0; i < 4; i++ { binary.Write(buf, binary.LittleEndian, uint32(0)) }

	binary.Write(buf, binary.LittleEndian, float64(1.0)) // Movement Multi
	binary.Write(buf, binary.LittleEndian, float64(1.0)) // Attack Multi
	binary.Write(buf, binary.LittleEndian, float64(8.0)) // Radius
	binary.Write(buf, binary.LittleEndian, float64(24.0)) // Height

	binary.Write(buf, binary.LittleEndian, uint32(char.HairStyle))
	binary.Write(buf, binary.LittleEndian, uint32(char.HairColor))
	binary.Write(buf, binary.LittleEndian, uint32(char.Face))
	binary.Write(buf, binary.LittleEndian, uint32(0)) // IsGM

	buf.Write(encodeUTF16(char.Title))

	// Клановая часть (исправлена под С1)
	binary.Write(buf, binary.LittleEndian, uint32(0)) // ClanID
	binary.Write(buf, binary.LittleEndian, uint32(0)) // CrestID
	binary.Write(buf, binary.LittleEndian, uint32(0)) // AllyID
	binary.Write(buf, binary.LittleEndian, uint32(0)) // AllyCrestID
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Режим (Sit/Stand/Ride)
	
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Private Store (В С1 это D, а не байт)
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Crafter (D)
	binary.Write(buf, binary.LittleEndian, uint32(0)) // PK (D)
	binary.Write(buf, binary.LittleEndian, uint32(0)) // PVP (D)
	binary.Write(buf, binary.LittleEndian, uint16(0)) // Cubic Count
	binary.Write(buf, binary.LittleEndian, uint16(0)) // Find Party (H)
	
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Invisible (D)
	binary.Write(buf, binary.LittleEndian, uint32(0)) // ? (D)
	binary.Write(buf, binary.LittleEndian, uint32(0)) // Clan Privs

	// ФИНАЛЬНЫЙ БЛОК: 7 штук WriteD(0)
	for i := 0; i < 7; i++ { 
		binary.Write(buf, binary.LittleEndian, uint32(0)) 
	}
	
	// Рекомендации (2 по 2 байта)
	binary.Write(buf, binary.LittleEndian, uint16(0)) // RecRemain
	binary.Write(buf, binary.LittleEndian, uint16(0)) // EvalScore

	return buf.Bytes()
}

func PackItemList(items []db.ItemData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x27)

	binary.Write(buf, binary.LittleEndian, uint16(1)) 
	binary.Write(buf, binary.LittleEndian, uint16(len(items)))

	for _, it := range items {
		binary.Write(buf, binary.LittleEndian, uint16(0)) // type1
		binary.Write(buf, binary.LittleEndian, uint32(it.ObjectID))
		binary.Write(buf, binary.LittleEndian, uint32(it.ItemID))
		binary.Write(buf, binary.LittleEndian, uint32(it.Count))
		binary.Write(buf, binary.LittleEndian, uint16(0)) // type2
		binary.Write(buf, binary.LittleEndian, uint16(0xFF)) // padding

		// ВАЖНО: В С1 тут должен быть флаг надетости (1 или 0)
		isEquipped := uint16(0)
		if it.Loc == "PAPERDOLL" {
			isEquipped = 1
		}
		binary.Write(buf, binary.LittleEndian, isEquipped) 

		// КЛЮЧЕВОЕ ИСПРАВЛЕНИЕ: Бит-маска слота (BodyPart)
		// Без этого клиент не знает, в какой слот "положить" иконку
		var bodyPart uint32 = 0
		if it.Loc == "PAPERDOLL" {
			bodyPart = getBodyPartByID(it.ItemID)
		}
		binary.Write(buf, binary.LittleEndian, bodyPart) 

		binary.Write(buf, binary.LittleEndian, uint16(it.EnchantLevel))
		binary.Write(buf, binary.LittleEndian, uint16(0)) // padding
	}
	return buf.Bytes()
}

func getBodyPartByID(itemID int32) uint32 {
	switch itemID {
	case 1, 6, 10, 2369, 2370: return 0x80  // Right Hand (Weapon)
	case 1146, 425:           return 0x400 // Chest
	case 1147, 461:           return 0x800 // Legs
	case 2368:                return 0x200 // Gloves
	case 2453:                return 0x1000 // Feet
	}
	return 0
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

// 0x48 - ValidateLocation (Синхронизация координат)
func PackValidateLocation(char *db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x48)
	binary.Write(buf, binary.LittleEndian, char.ObjectID)
	binary.Write(buf, binary.LittleEndian, int32(char.X))
	binary.Write(buf, binary.LittleEndian, int32(char.Y))
	binary.Write(buf, binary.LittleEndian, int32(char.Z))
	binary.Write(buf, binary.LittleEndian, int32(0)) // Heading или 0
	return buf.Bytes()
}

// 0x2A - EquipItemSuccess
func PackEquipItemSuccess(slot int32) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x2A)
	binary.Write(buf, binary.LittleEndian, slot)
	return buf.Bytes()
}

// 0x0E - StatusUpdate
func PackStatusUpdate(char *db.CharData) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(0x0E)
	binary.Write(buf, binary.LittleEndian, char.ObjectID)
	
	// Количество атрибутов, которые обновляем (например, 2: HP и MP)
	binary.Write(buf, binary.LittleEndian, uint32(2)) 
	
	// 0x09 - CUR_HP (из твоего списка)
	binary.Write(buf, binary.LittleEndian, uint32(0x09))
	binary.Write(buf, binary.LittleEndian, uint32(char.CurHp))
	
	// 0x0B - CUR_MP (из твоего списка)
	binary.Write(buf, binary.LittleEndian, uint32(0x0B))
	binary.Write(buf, binary.LittleEndian, uint32(char.CurMp))
	
	return buf.Bytes()
}
