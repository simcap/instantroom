package client

import (
	"crypto/aes"
	"crypto/cipher"
	"github.com/gorilla/websocket"
	"log"
)

type AESConnection struct {
	conn  *websocket.Conn
	block cipher.Block
}

func NewAESConnection(c *websocket.Conn, key []byte) *AESConnection {
	b, err := aes.NewCipher(key)
	if err != nil {
		log.Fatalf("Failed to build block cypher: %s", err)
	}
	return &AESConnection{c, b}
}

func (c *AESConnection) EncryptMessage(msg []byte) error {
	var dst = make([]byte, 64)
	c.block.Encrypt(dst, msg)
	return c.conn.WriteMessage(websocket.BinaryMessage, dst)
}

func (c *AESConnection) DecryptMessage() ([]byte, error) {
	_, msg, err := c.conn.ReadMessage()
	var dst = make([]byte, 64)
	c.block.Decrypt(dst, msg)
	return dst, err
}
