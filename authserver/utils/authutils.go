package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/scrypt"
)

var SECRET []byte = []byte(os.Getenv("SECRET_KEY"))
var ACCESS []byte = []byte(os.Getenv("ACCESS_KEY"))

// generate crypto random for tokens and PW Salt.
func randomCryptoBytes() ([]byte, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Cryptographic string for token use.
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

	pwHashBytes, err := hasher(salt, password)
	if err != nil {
		fmt.Println(err)
		return false, err

	}

	pwHashed := base64.StdEncoding.EncodeToString(pwHashBytes)

	if len(pwHashed) != len(storedHashStr) {
		fmt.Println("Hash not same len")
		return false, nil
	}

	if pwHashed != storedHashStr {
		fmt.Println("Hash Mismatch")
		return false, nil
	}

	return true, nil
}

func GetPasswordHash(plainTxtPW string) string {
	salt, _ := randomCryptoBytes()
	hash, _ := hasher(salt, plainTxtPW)
	return base64.StdEncoding.EncodeToString(hash)
}

//=========================================//
// ---- JWT Creation and Verification ---- //
//=========================================//

var jwtHeader map[string]string = map[string]string{
	"alg": "HS256",
	"typ": "JWT",
}

type TokenClaims struct {
	User_id  int       `json:"id"`
	Username string    `json:"username"`
	Is_staff bool      `json:"is_staff"`
	Exp      time.Time `json:"exp"`
}

func base64Encode(src []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(src), "=")
}

// Generate new Access JWT Token
func GenerateAccessToken(claims *TokenClaims) (string, error) {
	privKey, err := loadRSAPrivateKey()
	if err != nil {
		return "", err
	}

	headerJSON, _ := json.Marshal(jwtHeader)
	payloadJSON, _ := json.Marshal(claims)
	headerEnc := base64Encode(headerJSON)
	payloadEnc := base64Encode(payloadJSON)
	head_payload := fmt.Sprintf("%s.%s", headerEnc, payloadEnc)

	sha := sha256.New()
	sha.Write([]byte(head_payload))
	signer, _ := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, sha.Sum(nil))
	signerEnc := base64Encode(signer)
	return fmt.Sprintf("%s.%s", head_payload, signerEnc), nil
}

// Verify JWT
// Returns Payload if no errors while decoding and signature matches
// Returns 'Expired' Error if expired
func ValidateAccessToken(jwt string) (*TokenClaims, error) {
	var header map[string]string
	var payload TokenClaims

	rsaPub, err := loadRSAPublicKey()
	if err != nil {
		return nil, err
	}

	token := strings.Split(jwt, ".")
	signer := token[2]

	head_payload := fmt.Sprintf("%s.%s", token[0], token[1])
	// mac := hmac.New(sha256.New, ACCESS)
	// mac.Write([]byte(head_payload))
	// sigCheck := base64Encode(mac.Sum(nil))

	// if signer != sigCheck {
	// 	return nil, fmt.Errorf("signature does not match")
	// }
	signerDec, err := base64.URLEncoding.DecodeString(signer)
	if err != nil {
		fmt.Println("signature Decoding failed")
		return nil, err
	}

	sha := sha256.New()
	sha.Write([]byte(head_payload))

	verifyErr := rsa.VerifyPKCS1v15(rsaPub, crypto.SHA256, sha.Sum(nil), signerDec)
	if verifyErr != nil {
		return nil, verifyErr
	}

	headerDec, _ := base64.RawURLEncoding.DecodeString(token[0])
	json.Unmarshal(headerDec, &header)
	if header["alg"] != jwtHeader["alg"] {
		return nil, fmt.Errorf("invalid algorithm")
	}

	payloadDec, _ := base64.RawStdEncoding.DecodeString(token[1])
	json.Unmarshal(payloadDec, &payload)

	if payload.Exp.Before(time.Now().UTC()) {
		return &TokenClaims{}, fmt.Errorf("expired")
	}

	return &payload, nil
}

//================================//
// ---- RSA Key File Loaders ---- //
//================================//

func loadRSAPrivateKey() (*rsa.PrivateKey, error) {
	// wd, _ := os.Getwd()
	// path := fmt.Sprintf("%s/../id_rsa", wd)
	privateKeyFile, err := os.ReadFile("signing_key.pem")
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(privateKeyFile)
	if block == nil {
		return nil, fmt.Errorf("failed to parse private Key")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not a pem format")
	}

	return privateKey, nil
}

func loadRSAPublicKey() (*rsa.PublicKey, error) {
	// wd, _ := os.Getwd()
	// path := fmt.Sprintf("%s/../id_rsa.pub", wd)
	publicKeyFile, err := os.ReadFile("public_key.pem")
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(publicKeyFile)
	if block == nil {
		return nil, fmt.Errorf("failed to parse public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if publicKey == nil || err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPub, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not RSA public key")
	}

	return rsaPub, nil
}
