package crypto_rsa

import (
	"Peerster/fail"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"os"
)

const (
	privateKeyBitSize = 2048 //should not be modified (some [256]byte variables depend on it (256bytes=2048bits))
)

func GeneratePrivateKey() *rsa.PrivateKey {
	reader := rand.Reader

	key, err := rsa.GenerateKey(reader, privateKeyBitSize)
	fail.HandleError(err)

	return key
}

func PrivateKeyToBytes(key *rsa.PrivateKey) []byte {
	return x509.MarshalPKCS1PrivateKey(key)
}

func PublicKeyToBytes(key *rsa.PublicKey) ([]byte, error) {
	return asn1.Marshal(key)
}

func BytesToPublicKey(bytes []byte) (*rsa.PublicKey, error) {
	var pubKey rsa.PublicKey
	_, err := asn1.Unmarshal(bytes, &pubKey)
	return &pubKey, err
}

func SavePEMKey(fileName string, key *rsa.PrivateKey) {
	bytes := PrivateKeyToBytes(key)

	err := SaveAsPEMKey(fileName, "PRIVATE KEY", bytes)
	fail.HandleError(err)

}

func SavePublicPEMKey(fileName string, pubKey *rsa.PublicKey) {
	bytes, err := PublicKeyToBytes(pubKey)
	fail.HandleError(err)

	err = SaveAsPEMKey(fileName, "PUBLIC KEY", bytes)
	fail.HandleError(err)
}

func SaveAsPEMKey(fileName, type_ string, bytes []byte) error {
	var pemkey = &pem.Block{
		Type:  type_,
		Bytes: bytes,
	}

	pemfile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer fail.HandleError(pemfile.Close())

	return pem.Encode(pemfile, pemkey)
}
