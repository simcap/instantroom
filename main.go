package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"log"
	"math/big"
	"net/http"
	"strings"

	redis "github.com/xuyu/goredis"
)

func main() {
	http.HandleFunc("/room", room)
	http.HandleFunc("/join", join)
	http.ListenAndServe(":8080", nil)
}

func room(w http.ResponseWriter, r *http.Request) {
	client, err := redis.Dial(&redis.DialConfig{Address: "127.0.0.1:6379"})
	if err != nil {
		log.Printf("Error connecting to redis. %s", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
	room := r.FormValue("room")
	pkey := r.FormValue("pkey")

	keyerr := client.Set(room, pkey, 3600, 0, false, true)
	if keyerr != nil {
		log.Printf("Cannot create room '%s'. %s", room, keyerr)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func join(w http.ResponseWriter, r *http.Request) {
	client, err := redis.Dial(&redis.DialConfig{Address: "127.0.0.1:6379"})

	room := r.FormValue("room")
	sigints := strings.Split(r.FormValue("sig"), ",")

	pkey, err := client.Get(room)
	pkeyder, _ := base64.StdEncoding.DecodeString(string(pkey))

	public_key, err := x509.ParsePKIXPublicKey(pkeyder)
	if err != nil {
		log.Printf("Cannot decode public key for room '%s'. %s", room, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	x, _ := new(big.Int).SetString(sigints[0], 10)
	y, _ := new(big.Int).SetString(sigints[1], 10)

	switch public_key := public_key.(type) {
	case *ecdsa.PublicKey:
		validsig := ecdsa.Verify(public_key, []byte("secured"), x, y)
		if validsig {
			log.Printf("Signature valid for %s", room)
			http.Error(w, "", http.StatusOK)
		} else {
			http.Error(w, "", http.StatusInternalServerError)
		}
		return
	default:
		log.Printf("Fail challenge for joinig room '%s'", room)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

}
