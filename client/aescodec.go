package client

import (
	"crypto/aes"
	"errors"
	"log"

	"golang.org/x/net/websocket"
)

func NewAESCodec(key []byte) websocket.Codec {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("New cipher with given failed: %s", err)
	}

	aesEncrypt := func(v interface{}) (msg []byte, payloadType byte, err error) {
		switch data := v.(type) {
		case []byte:
			var dst = make([]byte, 64)
			block.Encrypt(dst, data)
			return dst, websocket.BinaryFrame, nil
		default:
			log.Fatalf("encrypt: []byte type expected: got %v", data)
		}
		return nil, websocket.BinaryFrame, errors.New("encrypt failed")
	}

	aesDecrypt := func(msg []byte, payloadType byte, v interface{}) (err error) {
		switch data := v.(type) {
		case *[]byte:
			var dst = make([]byte, 64)
			block.Decrypt(dst, msg)
			*data = dst
			return nil
		default:
			log.Fatalf("decrypt: *[]byte type expected: got %v", data)
		}
		return nil
	}

	return websocket.Codec{aesEncrypt, aesDecrypt}
}
