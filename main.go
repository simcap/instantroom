package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"

	redislib "github.com/xuyu/goredis"
)

var redis = initRedis()
var challenge = []byte("secured")

func initRedis() *redislib.Redis {
	client, err := redislib.Dial(&redislib.DialConfig{Address: "127.0.0.1:6379"})
	if err != nil {
		log.Printf("Cannot connect to redis. %s", err)
		os.Exit(1)
	}
	log.Print("Connected to redis")
	return client
}

func main() {
	http.HandleFunc("/room", room)
	http.HandleFunc("/join", join)
	http.ListenAndServe(":8080", nil)
}

func room(w http.ResponseWriter, r *http.Request) {
	room := r.FormValue("room")
	pkey := r.FormValue("pkey")

	keyerr := redis.Set(room, pkey, 3600, 0, false, true)
	if keyerr != nil {
		log.Printf("Cannot create room '%s'. %s", room, keyerr)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func join(w http.ResponseWriter, r *http.Request) {
	room := r.FormValue("room")
	sigints := strings.Split(r.FormValue("sig"), ",")

	pubkey, err := getPublicKey(room)
	if err != nil {
		log.Printf("Cannot use public key for room '%s'. %s", room, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	x, _ := new(big.Int).SetString(sigints[0], 10)
	y, _ := new(big.Int).SetString(sigints[1], 10)

	validsig := ecdsa.Verify(pubkey, challenge, x, y)
	if validsig {
		log.Printf("Joining '%s': signature valid", room)
		return
	} else {
		http.Error(w, "", http.StatusNotFound)
	}
}

func getPublicKey(room string) (*ecdsa.PublicKey, error) {
	key, err := redis.Get(room)
	if err != nil {
		m := fmt.Sprintf("Public key for room '%s': error retrieving key: %s", room, err)
		log.Print(m)
		return nil, errors.New(m)
	}

	keyder, err := base64.StdEncoding.DecodeString(string(key))
	if err != nil {
		m := fmt.Sprintf("Public key for room '%s': failed to base64 decode: %s", room, err)
		log.Print(m)
		return nil, errors.New(m)
	}

	pubkey, err := x509.ParsePKIXPublicKey(keyder)
	switch pubkey := pubkey.(type) {
	case *ecdsa.PublicKey:
		return pubkey, nil
	default:
		return nil, fmt.Errorf("Public key for room '%s': invalid type (%s)", room, pubkey)
	}
}
