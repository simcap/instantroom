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
	var AES = client.NewAESCodec(aeskeybytes)

	err := user.CreateRoom(room)
	if err != nil {
		t.Errorf("Cannot create room %s: %s", room, err)
	}

	ws, err := user.JoinRoom(room)
	if err != nil {
		t.Errorf("Websocket connection failed for room '%s'. %s", room, err)
	}

	ws2, err := user2.JoinRoom(room)
	if err != nil {
		t.Errorf("Websocket connection failed for room '%s'. %s", room, err)
	}
	ws2 = ws2

	if err := AES.Send(ws, []byte("hello here I am mister")); err != nil {
		t.Errorf("Failed to write message to correctly open websocket: %s", err)
	}

	var msg []byte
	if err := AES.Receive(ws, &msg); err != nil {
		t.Errorf("Failed to read message '%s' to correctly open websocket: %s", string(msg), err)
	}

	fmt.Printf("Received message: %q\n", string(msg))
}
