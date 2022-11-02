package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"log"

	"golang.org/x/crypto/nacl/box"
)

const (
	GH_KEY    string = "GH_KEY"
	GH_KEY_ID string = "GH_KEY_ID"
)

type Secret struct {
	KeyID string `json:"key_id"`
	Value string `json:"encrypted_value"`
}

func main() {
	flag.Usage = func() {
		cmd := filepath.Base(os.Args[0])
		fmt.Printf("%s [-key ENV_NAME][-key-id ENV_NAME] value\n", cmd)
		flag.PrintDefaults()
	}
	ghKeyEnv := flag.String("key", GH_KEY, "name of the env var")
	ghKeyIDenv := flag.String("key-id", GH_KEY_ID, "name of the env var")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Println("value env var name argument is required!")
		os.Exit(1)
	}

	value, ok := os.LookupEnv(flag.Arg(0))
	if !ok {
		log.Printf("provided env var \"%s\" with value not found\n", flag.Arg(0))
		os.Exit(1)
	}

	base64key, ok := os.LookupEnv(*ghKeyEnv)
	if !ok {
		log.Printf("required env var \"%s\" with key data not found\n", *ghKeyEnv)
		os.Exit(1)
	}

	keyID, ok := os.LookupEnv(*ghKeyIDenv)
	if !ok {
		log.Printf("required env var \"%s\" with key ID not found\n", *ghKeyIDenv)
		os.Exit(1)
	}

	gh_key, err := base64.StdEncoding.DecodeString(base64key)
	if err != nil {
		log.Println("key did not decode (base64)")
		os.Exit(1)
	}

	var secretKey [32]byte
	copy(secretKey[:], gh_key)

	encMsg := []byte{}
	encMsg, err = box.SealAnonymous(encMsg, []byte(value), &secretKey, rand.Reader)
	if err != nil {
		log.Printf("encrypt didn't work: %s\n", err)
		os.Exit(1)
	}
	// print the result to stdout, base64 encoded
	secret := Secret{keyID, base64.StdEncoding.EncodeToString(encMsg)}
	// fmt.Println(base64.StdEncoding.EncodeToString(encMsg))
	data, err := json.Marshal(secret)
	if err != nil {
		log.Printf("couldn't encode secret: %s\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}
