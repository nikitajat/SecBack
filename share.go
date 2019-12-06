package main

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/chinmay/sss"
)

type ShareID string

type Share struct {
	SID  ShareID
	Data []byte
}


func CreateShares(secret []byte, sid ShareID, n int) []Share {
	if n > 255 {
		panic("n > 255 not supported")
	}

	sharesBytes, err := sss.Split(byte(n), byte(n), secret)
	check(err)

	shares := make([]Share, n)
	i := 0
	for x, y := range sharesBytes {
		shares[i].SID = sid
		shares[i].Data = append(y, x)
		i++
	}

	return shares
}


func CombineShares(shares []Share) []byte {
	if len(shares) > 255 {
		panic("n > 255 not supported")
	}

	sharesBytes := make(map[byte][]byte)
	for _, v := range shares {
		i := v.Data[len(v.Data)-1]
		sharesBytes[i] = v.Data[:len(v.Data)-1]
	}

	return sss.Combine(sharesBytes)
}

func RandomShareID() ShareID {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}

	return ShareID(base64.URLEncoding.EncodeToString(randomBytes))
}
