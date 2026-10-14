package coinglass

import (
	"bytes"
	"crypto/aes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"time"
)

// generateLiqSign 生成清算类接口所需的 data 签名参数
// 🔐 还原自前端，清算地图（liqMap）与清算热力图（liqHeatMap）共用
//
// 签名流程:
//  1. ⏱️ 以 LiqSignTOTPSecret 计算标准 TOTP（HMAC-SHA1、步长30秒、6位）
//  2. 🧩 拼成 "<unix秒>,<6位码>"
//  3. 🔒 以 LiqSignAESKey 做 AES-256-ECB + PKCS7 加密
//  4. 📦 base64 编码
//
// 参数:
//
//	now - ⏰ 签名使用的时间（正常传 time.Now()，传入固定值便于离线验证）
func generateLiqSign(now time.Time) (string, error) {
	// 🔑 第一步：解码 base32 密钥并计算 TOTP
	secret, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(LiqSignTOTPSecret)
	if err != nil {
		return "", fmt.Errorf("❌ TOTP密钥解码失败: %w", err)
	}

	unix := now.Unix()
	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], uint64(unix/LiqSignTOTPStep))

	mac := hmac.New(sha1.New, secret)
	mac.Write(counter[:])
	sum := mac.Sum(nil)

	// 📐 RFC 4226 动态截断
	offset := sum[len(sum)-1] & 0x0f
	code := (binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff) % 1000000

	// 🧩 第二步：拼接明文
	plaintext := []byte(fmt.Sprintf("%d,%06d", unix, code))

	// 🔒 第三步：AES-256-ECB + PKCS7
	block, err := aes.NewCipher([]byte(LiqSignAESKey))
	if err != nil {
		return "", fmt.Errorf("❌ 创建AES加密器失败: %w", err)
	}

	padLen := aes.BlockSize - len(plaintext)%aes.BlockSize
	plaintext = append(plaintext, bytes.Repeat([]byte{byte(padLen)}, padLen)...)

	ciphertext := make([]byte, len(plaintext))
	for i := 0; i < len(plaintext); i += aes.BlockSize {
		block.Encrypt(ciphertext[i:i+aes.BlockSize], plaintext[i:i+aes.BlockSize])
	}

	// 📦 第四步：base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}
