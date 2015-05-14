package main

import (
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

	if _, err := ws.Write([]byte("hello")); err != nil {
		t.Errorf("Failed to write message to correctly open websocket: %s", err)
	}

	fmt.Printf("Reading %v", ws2)

	var msg = make([]byte, 100)
	if _, err := ws.Read(msg); err != nil {
		t.Errorf("Failed to read message '%s' to correctly open websocket: %s", string(msg), err)
	}

	fmt.Printf("Received message: %s", string(msg))
}
