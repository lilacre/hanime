package downloader

import (
	"crypto/aes"
	"crypto/cipher"
)

// Decrypt encrypted ts file
func decrypt(src []byte, iv string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(iv))
	if err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCDecrypter(block, []byte(iv))
	dstData := make([]byte, len(src))
	blockMode.CryptBlocks(dstData, src)
	return dstData, nil
}
