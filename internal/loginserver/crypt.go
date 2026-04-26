// darkages/internal/loginserver/crypt.go
package loginserver

import (
	"golang.org/x/crypto/blowfish"
)

type Crypt struct {
	cipher *blowfish.Cipher
}

func NewCrypt() *Crypt {
	// Твой токен (20 байт)
	baseKey := []byte{0x5b, 0x3b, 0x27, 0x2e, 0x5d, 0x39, 0x34, 0x2d, 0x33, 0x31, 0x3d, 0x3d, 0x2d, 0x25, 0x26, 0x40, 0x21, 0x5e, 0x2b, 0x5d}
	
	// Добавляем 21-й байт (ноль), как ты и писал
	realKey := append(baseKey, 0x00)
	
	c, _ := blowfish.NewCipher(realKey)
	return &Crypt{cipher: c}
}

func (c *Crypt) Decrypt(data []byte) {
	for i := 0; i < len(data); i += 8 {
		// 1. Разворачиваем LittleEndian в BigEndian для библиотеки Go
		data[i+0], data[i+3] = data[i+3], data[i+0]
		data[i+1], data[i+2] = data[i+2], data[i+1]
		data[i+4], data[i+7] = data[i+7], data[i+4]
		data[i+5], data[i+6] = data[i+6], data[i+5]

		// 2. Дешифруем 8 байт
		c.cipher.Decrypt(data[i:i+8], data[i:i+8])

		// 3. Разворачиваем обратно
		data[i+0], data[i+3] = data[i+3], data[i+0]
		data[i+1], data[i+2] = data[i+2], data[i+1]
		data[i+4], data[i+7] = data[i+7], data[i+4]
		data[i+5], data[i+6] = data[i+6], data[i+5]
	}
}

func (c *Crypt) Encrypt(data []byte) {
	for i := 0; i < len(data); i += 8 {
		// 1. Разворачиваем в BigEndian
		data[i+0], data[i+3] = data[i+3], data[i+0]
		data[i+1], data[i+2] = data[i+2], data[i+1]
		data[i+4], data[i+7] = data[i+7], data[i+4]
		data[i+5], data[i+6] = data[i+6], data[i+5]

		// 2. Шифруем
		c.cipher.Encrypt(data[i:i+8], data[i:i+8])

		// 3. Разворачиваем в LittleEndian
		data[i+0], data[i+3] = data[i+3], data[i+0]
		data[i+1], data[i+2] = data[i+2], data[i+1]
		data[i+4], data[i+7] = data[i+7], data[i+4]
		data[i+5], data[i+6] = data[i+6], data[i+5]
	}
}
