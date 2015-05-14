package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/simcap/instantroom/client"
)

var storagePath = filepath.Join(os.Getenv("HOME"), ".instantroom")

func main() {
	room := flag.String("r", "", "name of room to create")
	username := flag.String("u", "", "user name")
	join := flag.String("j", "", "name of room to join")
	host := flag.String("h", "127.0.0.1:8080", "address and port of host")

	flag.Parse()

	service := &client.Client{
		&client.CLIKeystore{storagePath},
		*username,
		*host,
	}

	if *room != "" {
		err := service.CreateRoom(*room)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *join != "" {
		privkey, _ := service.GetPrivateKey(*join)
		aeskeyhex := fmt.Sprintf("%x", privkey.D)
		aeskeybytes, _ := hex.DecodeString(aeskeyhex)
		var AES = client.NewAESCodec(aeskeybytes)

		ws, err := service.JoinRoom(*join)
		if err != nil {
			log.Fatal(err)
		}
		if err := AES.Send(ws, []byte("hello")); err != nil {
			log.Fatalf("Failed to write message on websocket: %s", err)
		}
		var msg = make([]byte, 100)
		if err := AES.Receive(ws, &msg); err != nil {
			log.Fatalf("Failed to read message on websocket: %s", err)
		}
		log.Printf("Message received: %s", string(msg))
	}
}
