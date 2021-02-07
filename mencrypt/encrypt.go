package mencrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
)

// key 16
// AES-128
// CBC
// PKCS7Padding

// expandKey 扩充key长度
func expandKey(key string) []byte {
	k := []byte(key)
	// 16, 24, 32
	if len(k) < 16 {
		k = append(k, bytes.Repeat([]byte{0}, 16-len(k))...)
	} else if len(k) < 24 {
		k = append(k, bytes.Repeat([]byte{0}, 24-len(k))...)
	} else if len(k) < 32 {
		k = append(k, bytes.Repeat([]byte{0}, 32-len(k))...)
	} else {
		k = k[:32]
	}
	return k
}

// AesEncrypt 加密
func AesEncrypt(orig string, key string) (string, error) {
	// 转成字节数组
	origData := []byte(orig)
	k := expandKey(key)
	// 分组秘钥
	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 补全码
	origData = PKCS7Padding(origData, blockSize)
	// 加密模式
	blockMode := cipher.NewCBCEncrypter(block, k[:blockSize])
	// 创建数组
	encrypted := make([]byte, len(origData))
	// 加密
	blockMode.CryptBlocks(encrypted, origData)

	return base64.StdEncoding.EncodeToString(encrypted), nil

}

// AesDecrypt aes解密
func AesDecrypt(encrypted string, key string) (string, error) {
	// 转成字节数组
	encryptedByte, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}
	k := expandKey(key)

	// 分组秘钥
	block, err := aes.NewCipher(k)
	if err != nil {
		return "", err
	}
	// 获取秘钥块的长度
	blockSize := block.BlockSize()
	// 加密模式
	blockMode := cipher.NewCBCDecrypter(block, k[:blockSize])
	// 创建数组
	orig := make([]byte, len(encryptedByte))
	// 解密
	blockMode.CryptBlocks(orig, encryptedByte)
	// 去补全码
	orig = PKCS7UnPadding(orig)
	return string(orig), nil
}

// PKCS7Padding 补码
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

// PKCS7UnPadding 去码
func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unPadding := int(origData[length-1])
	return origData[:(length - unPadding)]
}

// DecryptAesEcb aes ecb 解密
func DecryptAesEcb(data, key []byte) ([]byte, error) {
	cip, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypted := make([]byte, len(data))
	size := cip.BlockSize()

	for bs, be := 0, size; bs < len(data); bs, be = bs+size, be+size {
		cip.Decrypt(decrypted[bs:be], data[bs:be])
	}
	decrypted = PKCS7UnPadding(decrypted)
	return decrypted, nil
}
