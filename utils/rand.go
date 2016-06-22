package utils

import (
	"encoding/binary"
	"encoding/base64"
	"crypto/rand"
)

func GetRandInt64() int64 {
	bs := make([]byte, 8)
	_, err := rand.Read(bs)
	if err != nil {
		return GetRandInt64()
	}
	return int64(binary.BigEndian.Uint64(bs))
}


func GetRandString(size int) string {
	bs := make([]byte, size)
	_, err := rand.Read(bs)
	if err != nil {
		return GetRandString(size)
	}
	return base64.RawStdEncoding.EncodeToString(bs)
}

func GetRandBytes(size int) []byte{
	bs := make([]byte, size)
	_, err := rand.Read(bs)
	if err != nil {
		return GetRandBytes(size)
	}
	return bs
}
