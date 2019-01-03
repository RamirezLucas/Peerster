package utils

import (
	"Peerster/fail"
	"encoding/hex"
	"fmt"
	"math/rand"
)

func HashToHex(hash []byte) string {
	return hex.EncodeToString(hash)
}

func HexToHash(hexHash string) []byte {
	hash, err := hex.DecodeString(hexHash)
	if err != nil {
		fail.HandleAbort(fmt.Sprint("could not decode hexadecimal string '%s'", hexHash), err)
		return nil
	}
	return hash
}

func FirstNZero(n int, bytes []byte) bool {
	if n < 0 || n > len(bytes) {
		fail.HandleError(fmt.Errorf("FirstNZero failed with n=%d and len(bytes)=%d", n, len(bytes)))
		return false
	}
	return AllZero(bytes[:n])
}

func AllZero(bytes []byte) bool {
	for _, v := range bytes {
		if v != 0 {
			return false
		}
	}
	return true
}

func Random32Bytes() [32]byte {
	bytes := [32]byte{}
	rand.Read(bytes[:])
	return bytes
}
