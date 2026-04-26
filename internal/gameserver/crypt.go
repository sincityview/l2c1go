// darkages/internal/gameserver/crypt.go
package gameserver

import (
	"encoding/binary"
)

type GameCrypt struct {
	inKey  []byte
	outKey []byte
}

func NewGameCrypt(firstKey []byte) *GameCrypt {
	c := &GameCrypt{
		inKey:  make([]byte, 8),
		outKey: make([]byte, 8),
	}
	// Инициализация ключей как в процедуре Key() из твоего текста
	copy(c.inKey[0:4], firstKey)
	c.inKey[4] = 0xA1
	c.inKey[5] = 0x6C
	c.inKey[6] = 0x54
	c.inKey[7] = 0x87
	
	// key_cs = key_sc
	copy(c.outKey, c.inKey)
	return c
}

func (c *GameCrypt) Decrypt(data []byte) {
	var prev byte = 0
	for i := 0; i < len(data); i++ {
		temp := data[i]
		// Формула из твоего Decode(): input[k] = i1 ^ key ^ i; i = i1;
		data[i] = temp ^ c.inKey[i%8] ^ prev
		prev = temp
	}
	c.updateKey(c.inKey, uint32(len(data)))
}

func (c *GameCrypt) Encrypt(data []byte) {
	var prev byte = 0
	for i := 0; i < len(data); i++ {
		// Формула из твоего Encrypt(): raw[i] = temp2 ^ key ^ temp; temp = raw[i];
		data[i] = data[i] ^ c.outKey[i%8] ^ prev
		prev = data[i]
	}
	c.updateKey(c.outKey, uint32(len(data)))
}

func (c *GameCrypt) updateKey(key []byte, size uint32) {
	old := binary.LittleEndian.Uint32(key[0:4])
	old += size
	binary.LittleEndian.PutUint32(key[0:4], old)
}

func AddChecksum(data []byte) []byte {
	if len(data) < 4 {
		// Если пакет совсем крошечный, добьем его до 4 байт
		pad := make([]byte, 4-len(data))
		data = append(data, pad...)
	}

	var chksum uint32
	for i := 0; i < len(data); i += 4 {
		// Проверка, чтобы не выйти за границы слайса
		if i+4 <= len(data) {
			ecx := binary.LittleEndian.Uint32(data[i : i+4])
			chksum ^= ecx
		}
	}
	res := make([]byte, 4)
	binary.LittleEndian.PutUint32(res, chksum)
	return append(data, res...)
}

