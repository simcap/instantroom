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

	resp, err = http.PostForm("http://127.0.0.1:8080/join", url.Values{
		"room": {room},
		"sig":  {fmt.Sprintf("%s,%s", r, s)},
	})

	if status := resp.StatusCode; status != 200 {
		t.Errorf("Challenge failed for room '%s'. Expecting status 200 but was %d", room, status)
	}
}
