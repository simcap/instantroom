package client

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gorilla/websocket"
)

type Keystore interface {
	GetPublicKeyBytes(room string) ([]byte, error)
	GetPublicKeyBase64(room string) (string, error)
	GetPrivateKey(room string) (*ecdsa.PrivateKey, error)
	GenerateKeys(room string) (*ecdsa.PrivateKey, error)
}

type Client struct {
	Keystore
	Username string
	Host     string
}

func (c *Client) CreateRoom(room string) error {
	roomurl := fmt.Sprintf("http://%s/room", c.Host)

	c.GenerateKeys(room)

	pubkey_base64, _ := c.GetPublicKeyBase64(room)
	resp, err := http.PostForm(roomurl, url.Values{
		"room":     {room},
		"username": {c.Username},
		"pkey":     {pubkey_base64},
	})

	if err != nil {
		return err
	}

	if status := resp.StatusCode; status != 200 {
		return fmt.Errorf("Failed to create room '%s' at %s", room, roomurl)
	}

	return nil
}

func (c *Client) JoinRoom(room string) (*websocket.Conn, error) {
	privkey, err := c.GetPrivateKey(room)
	if err != nil {
		return nil, fmt.Errorf("Joining room '%s': cannot get private key: %s", room, err)
	}

	r, s, err := ecdsa.Sign(rand.Reader, privkey, []byte("secured"))
	if err != nil {
		return nil, fmt.Errorf("Joining room '%s': cannot sign with private key: %s", room, err)
	}

	origin := fmt.Sprintf("http://%s/", c.Host)
	header := http.Header{"Origin": {origin}}
	u, _ := url.Parse(fmt.Sprintf("ws://%s/join", c.Host))
	params := url.Values{}
	params.Add("room", room)
	params.Add("username", c.Username)
	params.Add("sig", fmt.Sprintf("%s,%s", r, s))
	u.RawQuery = params.Encode()

	conn, _, errws := websocket.DefaultDialer.Dial(u.String(), header)
	if errws != nil {
		return nil, fmt.Errorf("Joining room '%s': websocket dial failed: %s", room, errws)
	}

	return conn, nil
}

// Keystore implementation for a command line client
type CLIKeystore struct {
	StoragePath string
}

func (k *CLIKeystore) GenerateKeys(room string) (*ecdsa.PrivateKey, error) {
	os.MkdirAll(filepath.Join(k.StoragePath, room), 0700)

	privkey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	privkey_bytes, _ := x509.MarshalECPrivateKey(privkey)
	pubkey_bytes, _ := x509.MarshalPKIXPublicKey(&privkey.PublicKey)

	ioutil.WriteFile(
		filepath.Join(k.StoragePath, room, "private.der"),
		privkey_bytes,
		0644,
	)
	ioutil.WriteFile(
		filepath.Join(k.StoragePath, room, "public.bin"),
		pubkey_bytes,
		0644,
	)

	return privkey, nil
}

func (k *CLIKeystore) GetPublicKeyBytes(room string) ([]byte, error) {
	return ioutil.ReadFile(filepath.Join(k.StoragePath, room, "public.bin"))
}

func (k *CLIKeystore) GetPublicKeyBase64(room string) (string, error) {
	pubkey, err := k.GetPublicKeyBytes(room)
	encoded := base64.StdEncoding.EncodeToString(pubkey)
	return encoded, err
}

func (k *CLIKeystore) GetPrivateKey(room string) (*ecdsa.PrivateKey, error) {
	privkey_bytes, _ := ioutil.ReadFile(filepath.Join(k.StoragePath, room, "private.der"))
	privkey, err := x509.ParseECPrivateKey(privkey_bytes)
	return privkey, err
}

// Keystore implementation for test
type MemoryKeystore struct {
	privkeys map[string]*ecdsa.PrivateKey
}

func NewMemoryKeystore() *MemoryKeystore {
	return &MemoryKeystore{make(map[string]*ecdsa.PrivateKey)}
}

func (k *MemoryKeystore) GetPublicKeyBytes(room string) ([]byte, error) {
	privkey, err := k.GetPrivateKey(room)
	if err != nil {
		return nil, err
	}
	bytes, err := x509.MarshalPKIXPublicKey(&privkey.PublicKey)
	return bytes, err
}

func (k *MemoryKeystore) GetPublicKeyBase64(room string) (string, error) {
	bytes, err := k.GetPublicKeyBytes(room)
	encoded := base64.StdEncoding.EncodeToString(bytes)
	return encoded, err
}

func (k *MemoryKeystore) GetPrivateKey(room string) (*ecdsa.PrivateKey, error) {
	return k.privkeys[room], nil
}

func (k *MemoryKeystore) GenerateKeys(room string) (*ecdsa.PrivateKey, error) {
	p, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	k.privkeys[room] = p
	return p, err
}
