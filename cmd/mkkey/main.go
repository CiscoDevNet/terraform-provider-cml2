package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"log"

	"golang.org/x/crypto/nacl/box"
)

const GH_KEY string = "GH_KEY"

func main() {
	flag.Usage = func() {
		cmd := filepath.Base(os.Args[0])
		fmt.Printf("%s [-key KEY_ENV_NAME] message\n", cmd)
		flag.PrintDefaults()
	}
	envVarName := flag.String("key", GH_KEY, "name of the env var")
	flag.Parse()
	if flag.NArg() == 0 {
		log.Println("message argument is required!")
		os.Exit(1)
	}
	message := flag.Arg(0)

	base64key, ok := os.LookupEnv(*envVarName)
	if !ok {
		log.Printf("required env var \"%s\" with key not found\n", *envVarName)
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
	encMsg, err = box.SealAnonymous(encMsg, []byte(message), &secretKey, rand.Reader)
	if err != nil {
		log.Printf("encrypt didn't work: %s\n", err)
		os.Exit(1)
	}
	// print the result to stdout, base64 encoded
	fmt.Println(base64.StdEncoding.EncodeToString(encMsg))
}
