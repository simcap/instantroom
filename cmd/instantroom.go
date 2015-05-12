package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"golang.org/x/net/websocket"
)

var storagePath = filepath.Join(os.Getenv("HOME"), ".instantroom")

func main() {
	room := flag.String("r", "", "name of room to create")
	username := flag.String("u", "", "user name")
	join := flag.String("j", "", "name of room to join")
	host := flag.String("h", "127.0.0.1:8080", "address and port of host")
	//send := flag.String("s", "ping", "message to be send to room")

	flag.Parse()

	os.MkdirAll(storagePath, 0700)

	if *room != "" {
		createRoom(*room, *host)
	}

	if *join != "" {
		joinRoom(*username, *join, *host)
	}
}

func createRoom(room string, host string) {
	os.Mkdir(filepath.Join(storagePath, room), 0700)

	privkey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privkey_bytes, _ := x509.MarshalECPrivateKey(privkey)
	pubkey_bytes, _ := x509.MarshalPKIXPublicKey(&privkey.PublicKey)

	ioutil.WriteFile(
		filepath.Join(storagePath, room, "private.der"),
		privkey_bytes,
		0644,
	)
	ioutil.WriteFile(
		filepath.Join(storagePath, room, "public.bin"),
		pubkey_bytes,
		0644,
	)

	resp, err := http.PostForm(fmt.Sprintf("http://%s/room", host), url.Values{
		"room": {room},
		"pkey": {base64.StdEncoding.EncodeToString(pubkey_bytes)},
	})

	if err != nil {
		fmt.Printf("Cannot post to host %s", err)
		os.Exit(1)
	}

	if status := resp.StatusCode; status == 200 {
		fmt.Printf("Created room %s successfully", room)
	} else {
		fmt.Printf("Uploading key to room '%s'. Expecting status 200 but was %d", room, status)
		os.Exit(1)
	}
}

func joinRoom(username string, room string, host string) {
	if username == "" || room == "" {
		fmt.Println("Username and room needed to join room")
		os.Exit(1)
	}

	privkey_bytes, _ := ioutil.ReadFile(filepath.Join(storagePath, room, "private.der"))
	privkey, err := x509.ParseECPrivateKey(privkey_bytes)

	r, s, err := ecdsa.Sign(rand.Reader, privkey, []byte("secured"))
	if err != nil {
		fmt.Printf("Cannot sign with private key. %s", err)
		os.Exit(1)
	}

	origin := fmt.Sprintf("http://%s/", host)
	u, _ := url.Parse(fmt.Sprintf("ws://%s/join", host))
	params := url.Values{}
	params.Add("room", room)
	params.Add("username", username)
	params.Add("sig", fmt.Sprintf("%s,%s", r, s))
	u.RawQuery = params.Encode()
	_, wserr := websocket.Dial(u.String(), "", origin)

	if wserr != nil {
		fmt.Printf("Websocket connection failed for room '%s'. %s", room, wserr)
		os.Exit(1)
	}
}
