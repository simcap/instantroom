package main

import (
	"testing"

	"github.com/simcap/instantroom/client"
)

func TestCreateRoomAndJoin(t *testing.T) {
	room := "biblical"
	service := &client.Client{
		client.NewMemoryKeystore(),
		"johntest",
		"127.0.0.1:8080",
	}

	service.GenerateKeys(room)

	err := service.CreateRoom(room)
	if err != nil {
		t.Errorf("Cannot create room %s: %s", room, err)
	}

	ws, err := service.JoinRoom(room)
	if err != nil {
		t.Errorf("Websocket connection failed for room '%s'. %s", room, err)
	}

	if _, err := ws.Write([]byte("hello\n")); err != nil {
		t.Errorf("Failed to write message to correctly open websocket")
	}
}
