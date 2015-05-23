package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

		ws, err := service.JoinRoom(*join)
		var AES = client.NewAESConnection(ws, aeskeybytes)

		if err != nil {
			log.Fatal(err)
		}

		go func() {
			for {
				msg, err := AES.DecryptMessage()
				if err == nil {
					splits := strings.Split(string(msg), ":")
					name, message := splits[0], splits[1]
					if name != *username {
						log.Printf("%s> %s", name, message)
					}
				}
			}
		}()

		for {
			message := ""
			_, errscan := fmt.Scanln(&message)
			if errscan != nil {
				log.Fatalf("Failed to read user input %s (message: '%q')", err, message)
			}
			if err := AES.EncryptMessage([]byte(*username + ":" + message)); err != nil {
				log.Fatalf("Failed to write message on websocket: %s", err)
			}
		}
	}
}
