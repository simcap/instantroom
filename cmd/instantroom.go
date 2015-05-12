package main

import (
	"flag"
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
		err := service.JoinRoom(*join)
		if err != nil {
			log.Fatal(err)
		}
	}
}
