package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/scrypt"
)

var SECRET []byte = []byte(os.Getenv("SECRET_KEY"))

// generate crypto random for Session ID as well PW Salt.
func randomCryptoBytes() ([]byte, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// For session_id use
func GenerateCryptoString() (string, error) {
	bytes, err := randomCryptoBytes()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hasher(salt []byte, plainTxtPW string) ([]byte, error) {
	passwordBytes := []byte(plainTxtPW)

	hash, err := scrypt.Key(passwordBytes, salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, err
	}
	passwordHash := append(salt, hash...)
	return passwordHash, nil
} // base64.StdEncoding.EncodeToString(passwordHash),

func VerifyPassword(storedHashStr, password string) (bool, error) {
	byteHash, err := base64.StdEncoding.DecodeString(storedHashStr)
	if err != nil {
		return false, err
	}
	salt := byteHash[:16]

	pwHashed, err := hasher(salt, password)
	if err != nil {
		fmt.Println(err)
		return false, err

	} else if len(pwHashed) != len(byteHash) {
		fmt.Println("Hash not same len")
		return false, nil
	}

	for i := range pwHashed {
		if pwHashed[i] != byteHash[i] {
			fmt.Println("Hash Byte mismatch")
			return false, nil
		}
	}
	return true, nil
}

func GetPasswordHash(plainTxtPW string) string {
	salt, _ := randomCryptoBytes()
	hash, _ := hasher(salt, plainTxtPW)
	return base64.StdEncoding.EncodeToString(hash)
}

var jwtHeader map[string]string = map[string]string{
	"alg": "HS256",
	"typ": "JWT",
}

type TokenClaims struct {
	User_id  int       `json:"id"`
	Username string    `json:"username"`
	Is_staff bool      `json:"is_staff"`
	IAT      time.Time `json:"iat"`
}

func base64Encode(src []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(src), "=")
}

func GenerateAccessToken(claims *TokenClaims) (string, error) {
	headerJSON, _ := json.Marshal(jwtHeader)
	payloadJSON, _ := json.Marshal(claims)
	headerEnc := base64Encode(headerJSON)
	payloadEnc := base64Encode(payloadJSON)
	head_payload := fmt.Sprintf("%s.%s", headerEnc, payloadEnc)

	mac := hmac.New(sha256.New, SECRET)
	mac.Write([]byte(head_payload))
	signer := base64Encode(mac.Sum(nil))
	return fmt.Sprintf("%s.%s", head_payload, signer), nil
}

func ValidateAccessToken(jwt string) (*TokenClaims, error) {
	var header map[string]string
	var payload TokenClaims

	token := strings.Split(jwt, ".")
	signer, _ := base64.RawStdEncoding.DecodeString(token[2])

	head_payload := fmt.Sprintf("%s.%s", token[0], token[1])
	mac := hmac.New(sha256.New, SECRET)
	mac.Write([]byte(head_payload))

	verified := hmac.Equal(mac.Sum(nil), signer)
	if !verified {
		return nil, fmt.Errorf("Signature does not match")
	}

	headerDec, _ := base64.RawStdEncoding.DecodeString(token[0])
	json.Unmarshal(headerDec, &header)
	if header["alg"] != jwtHeader["alg"] {
		return nil, fmt.Errorf("Invalid Algorithm")
	}

	payloadDec, _ := base64.RawStdEncoding.DecodeString(token[1])
	json.Unmarshal(payloadDec, &payload)

	return &payload, nil
}
