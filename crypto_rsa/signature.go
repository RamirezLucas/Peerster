package crypto_rsa

import (
	"Peerster/messages"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

func NewSignature(file *messages.File, key *rsa.PrivateKey) ([256]byte, error) {
	hash := file.Hash()
	return Sign(hash[:], key)
}

func Sign(toSign []byte, key *rsa.PrivateKey) ([256]byte, error) {
	// crypto/rand.Reader is a good source of entropy for blinding the RSA operation.
	rng := rand.Reader

	// Only small messages can be signed directly; thus the hash of a
	// message, rather than the message itself, is signed. This requires
	// that the hash function be collision resistant. SHA-256 is the
	// least-strong hash function that should be used for this at the time
	// of writing (2019).
	hashed := sha256.Sum256(toSign)

	// Since signing is a randomized function, cipher will be different each time.
	bytes, err := rsa.SignPKCS1v15(rng, key, crypto.SHA256, hashed[:])
	signature := [256]byte{}
	copy(signature[:], bytes)
	return signature, err
}

func Verify(whatWasSigned []byte, newSig [256]byte, pubKey *rsa.PublicKey) error {

	// Only small messages can be signed directly; thus the hash of a
	// message, rather than the message itself, is signed. This requires
	// that the hash function be collision resistant. SHA-256 is the
	// least-strong hash function that should be used for this at the time
	// of writing (2019).
	hashed := sha256.Sum256(whatWasSigned)

	err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hashed[:], newSig[:])
	// no error = correctly signed
	return err
}
