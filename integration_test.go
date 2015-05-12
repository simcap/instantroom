package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/net/websocket"
)

func TestCreateRoom(t *testing.T) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pub_bytes, _ := x509.MarshalPKIXPublicKey(&priv.PublicKey)

	pub_base64 := base64.StdEncoding.EncodeToString(pub_bytes)

	room := "biblical"

	fmt.Printf("Posting pubkey: %s", pub_base64)
	resp, err := http.PostForm("http://127.0.0.1:8080/room", url.Values{
		"room": {room},
		"pkey": {pub_base64},
	})

	if err != nil {
		t.Errorf("Cannot post to server %s", err)
	}

	if status := resp.StatusCode; status != 200 {
		t.Errorf("Uploading key to room '%s'. Expecting status 200 but was %d", room, status)
	}

	r, s, err := ecdsa.Sign(rand.Reader, priv, []byte("secured"))
	if err != nil {
		t.Errorf("Cannot sign with private key. %s", err)
	}

	origin := "http://127.0.0.1:8080/"
	u, _ := url.Parse("ws://127.0.0.1:8080/join")
	params := url.Values{}
	params.Add("room", room)
	params.Add("username", "siegfried")
	params.Add("sig", fmt.Sprintf("%s,%s", r, s))
	u.RawQuery = params.Encode()
	ws, err := websocket.Dial(u.String(), "", origin)

	if err != nil {
		t.Errorf("Websocket connection failed for room '%s'. %s", room, err)
	}

	if _, err := ws.Write([]byte("hello\n")); err != nil {
		t.Errorf("Failed to write message to correctly open websocket")
	}
}
