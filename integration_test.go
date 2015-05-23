package main

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/simcap/instantroom/client"
)

func TestCreateRoomAndJoin(t *testing.T) {
	room := "biblical"
	keystore := client.NewMemoryKeystore()
	keystore.GenerateKeys(room)

	user := &client.Client{
		keystore,
		"johntest",
		"127.0.0.1:8080",
	}

	user2 := &client.Client{
		keystore,
		"billtest",
		"127.0.0.1:8080",
	}

	privkey, _ := user.GetPrivateKey(room)
	aeskeyhex := fmt.Sprintf("%x", privkey.D)
	aeskeybytes, _ := hex.DecodeString(aeskeyhex)

	ws, err := user.Chat(room)
	if err != nil {
		t.Errorf("Websocket connection failed for room '%s'. %s", room, err)
	}

	_, errws := user2.Chat(room)
	if errws != nil {
		t.Errorf("Websocket connection failed for room '%s'. %s", room, errws)
	}

	var AESConn = client.NewAESConnection(ws, aeskeybytes)

	if err := AESConn.EncryptMessage([]byte("hello here I am mister")); err != nil {
		t.Errorf("Failed to write message to correctly open websocket: %s", err)
	}

	msg, err := AESConn.DecryptMessage()
	if err != nil {
		t.Errorf("Failed to read message '%s' to correctly open websocket: %s", string(msg), err)
	}

	fmt.Printf("Received message: %q\n", string(msg))
}
