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

	"github.com/gorilla/websocket"
	redislib "github.com/xuyu/goredis"
)

var redis = initRedis()
var challenge = []byte("secured")
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func initRedis() *redislib.Redis {
	client, err := redislib.Dial(&redislib.DialConfig{Address: "127.0.0.1:6379"})
	if err != nil {
		log.Printf("Cannot connect to redis. %s", err)
		os.Exit(1)
	}
	log.Print("Connected to redis")
	return client
}

type Room struct {
	name  string
	users map[string]*websocket.Conn
}

var rooms = map[string]*Room{}

func main() {
	http.HandleFunc("/room", room)
	http.HandleFunc("/join", join)
	http.ListenAndServe(":8080", nil)
}

func room(w http.ResponseWriter, r *http.Request) {
	room := r.FormValue("room")
	username := r.FormValue("username")
	pkey := r.FormValue("pkey")

	keyerr := redis.Set(room, pkey, 3600, 0, false, true)
	if keyerr != nil {
		log.Printf("Cannot create room '%s'. %s", room, keyerr)
		http.Error(w, "", http.StatusInternalServerError)
	}

	log.Printf("Created room '%s' for user '%s'", room, username)
}

func join(w http.ResponseWriter, r *http.Request) {
	room := r.FormValue("room")
	username := r.FormValue("username")
	sigints := strings.Split(r.FormValue("sig"), ",")

	pubkey, err := getPublicKey(room)
	if err != nil {
		m := fmt.Sprintf("Cannot use public key for room '%s'. %s", room, err)
		log.Print(m)
		http.Error(w, m, http.StatusInternalServerError)
	}

	x, _ := new(big.Int).SetString(sigints[0], 10)
	y, _ := new(big.Int).SetString(sigints[1], 10)

	validsig := ecdsa.Verify(pubkey, challenge, x, y)
	if validsig {
		log.Printf("Handshaked for room '%s': signature valid", room)
		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			m := fmt.Sprintf("Connection updgrae failed for room '%s': %s", room, err)
			log.Print(m)
			http.Error(w, m, http.StatusInternalServerError)
		}
		dispatch(conn, room, username)
	} else {
		m := fmt.Sprintf("Handshake failed for room '%s': invalid signature", room)
		log.Print(m)
		http.Error(w, m, http.StatusInternalServerError)
	}
}

func dispatch(conn *websocket.Conn, roomname string, username string) {
	room := AddUserToRoom(conn, roomname, username)
	log.Printf("... start dispatching for %#v", room.users)
	for {
		_, msg, err := conn.ReadMessage()
		if err == nil {
			for u, c := range room.users {
				err := c.WriteMessage(websocket.BinaryMessage, msg)
				if err != nil {
					log.Printf("Failed replicating message from %s to %s in room: %s", username, u, err)
				}
			}
		}
	}
}

func AddUserToRoom(conn *websocket.Conn, room string, username string) *Room {
	if r, ok := rooms[room]; ok {
		if _, ok := r.users[username]; ok {
			log.Printf("%s already in room %s", username, room)
		} else {
			log.Printf("Adding new user %s for room %s", username, room)
			r.users[username] = conn
		}
		return r
	} else {
		log.Printf("Adding new room %s for username %s", room, username)
		newRoom := Room{name: room, users: make(map[string]*websocket.Conn)}
		newRoom.users[username] = conn
		rooms[room] = &newRoom
		return &newRoom
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
	if err != nil {
		return nil, fmt.Errorf("Public key for room '%s': failed to parse der: %s", room, err)
	}
	switch pubkey := pubkey.(type) {
	case *ecdsa.PublicKey:
		return pubkey, nil
	default:
		return nil, fmt.Errorf("Public key for room '%s': invalid type (%s)", room, pubkey)
	}
}
