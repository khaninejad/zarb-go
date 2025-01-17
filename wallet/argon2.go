package wallet

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/argon2"
)

// Parameters are set based on the spec recommendation
// Read more here https://datatracker.ietf.org/doc/html/rfc9106#section-4
var (
	iterations  = uint32(1)
	memory      = uint32(2 * 1024 * 1024)
	parallelism = uint8(4)
)

type argon2Encrypter struct {
	passphrase string
}

func newArgon2Encrypter(passphrase string) *argon2Encrypter {
	return &argon2Encrypter{
		passphrase: passphrase,
	}
}
func (e *argon2Encrypter) encrypt(message string) encrypted {
	// Random salt
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	exitOnErr(err)

	cipherKey := e.cipherKey(e.passphrase, salt, iterations, memory, parallelism)

	// Using salt for Initialization Vector (IV)
	iv := salt
	d := aesCrypt([]byte(message), iv, cipherKey)

	// Generate the mac
	mac := sha256MAC(cipherKey[16:32], d)

	params := newParams()
	params.SetUint32("iterations", iterations)
	params.SetUint32("memory", memory)
	params.SetUint8("parallelism", parallelism)
	params.SetBytes("salt", salt)
	params.SetBytes("mac", mac)

	cipherText := base64.StdEncoding.EncodeToString(d)

	return encrypted{
		Method:     "ARGON2ID_AES-256-CTR_SHA256",
		Params:     params,
		CipherText: cipherText,
	}
}

func (e *argon2Encrypter) decrypt(ct encrypted) (string, error) {
	if ct.Method != "ARGON2ID_AES-256-CTR_SHA256" {
		return "", ErrInvalidPassword
	}

	salt := ct.Params.GetBytes("salt")
	mac := ct.Params.GetBytes("mac")
	iterations := ct.Params.GetUint32("iterations")
	memory := ct.Params.GetUint32("memory")
	parallelism := ct.Params.GetUint8("parallelism")

	cipherKey := e.cipherKey(e.passphrase, salt, iterations, memory, parallelism)
	d, err := base64.StdEncoding.DecodeString(ct.CipherText)
	exitOnErr(err)

	// Using MAC to check if the password is correct
	// https: //en.wikipedia.org/wiki/Authenticated_encryption#Encrypt-then-MAC_(EtM)
	if !safeCmp(mac, sha256MAC(cipherKey[16:32], d)) {
		return "", ErrInvalidPassword
	}

	text := aesCrypt(d, salt, cipherKey)
	return string(text), nil
}

func (e *argon2Encrypter) cipherKey(passphrase string, salt []byte, iterations, memory uint32, parallelism uint8) []byte {
	// Argon2 currently has three modes: data-dependent Argon2d, data-independent Argon2i, and a mix of the two, Argon2id.
	return argon2.IDKey([]byte(passphrase), salt, iterations, memory, parallelism, 32)
}
