package tests

import (
	"Peerster/crypto_rsa"
	"Peerster/utils"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestSignThenVerify(t *testing.T) {

	key := crypto_rsa.GeneratePrivateKey()
	randomBytes := [256]byte{}
	rand.Read(randomBytes[:])
	signature, err := crypto_rsa.Sign(randomBytes[:], key)

	assert.Nil(t, err, "no error when signing")

	isVerified := crypto_rsa.Verify(randomBytes[:], signature, &key.PublicKey)

	assert.NoError(t, isVerified, "signature of byte should be verified correctly")
}

func TestSignThenVerifySomethingElse(t *testing.T) {

	key := crypto_rsa.GeneratePrivateKey()
	randomBytes := [256]byte{}
	rand.Read(randomBytes[:])
	signature, err := crypto_rsa.Sign(randomBytes[:], key)

	assert.Nil(t, err, "no error when signing")

	newRandomBytes := [256]byte{}
	rand.Read(randomBytes[:])
	assert.False(t, randomBytes == newRandomBytes, "two random byte sequence should not be equal")

	isVerified := crypto_rsa.Verify(newRandomBytes[:], signature, &key.PublicKey)

	assert.Error(t, isVerified, "signature of new random bytes should not be verified correctly")
}

func TestRandomSignature(t *testing.T) {
	key := crypto_rsa.GeneratePrivateKey()
	someBytes := utils.Random32Bytes()
	signature, err := crypto_rsa.Sign(someBytes[:], key)

	assert.Nil(t, err, "no error when signing")

	randomBytes := [256]byte{}
	rand.Read(randomBytes[:])

	assert.False(t, signature == randomBytes, "a random signature should not be equal to the original")

	isVerified := crypto_rsa.Verify(someBytes[:], randomBytes, &key.PublicKey)

	assert.Error(t, isVerified, "random signature should not be able verify correctly")
}
