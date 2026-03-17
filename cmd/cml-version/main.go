package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	cmlclient "github.com/rschmied/gocmlclient/pkg/client"
)

func main() {
	var (
		address    string
		token      string
		username   string
		password   string
		skipVerify bool
	)

	flag.StringVar(&address, "address", os.Getenv("CML_ADDRESS"), "CML base URL, e.g. https://cml.example")
	flag.StringVar(&token, "token", os.Getenv("CML_TOKEN"), "CML JWT token")
	flag.StringVar(&username, "username", os.Getenv("CML_USERNAME"), "CML username")
	flag.StringVar(&password, "password", os.Getenv("CML_PASSWORD"), "CML password")
	flag.BoolVar(&skipVerify, "skip-verify", true, "Skip TLS verification")
	flag.Parse()

	if address == "" {
		fmt.Fprintln(os.Stderr, "missing -address or CML_ADDRESS")
		os.Exit(2)
	}

	opts := []cmlclient.Option{cmlclient.SkipReadyCheck()}
	if skipVerify {
		opts = append(opts, cmlclient.WithInsecureTLS())
	}
	if token != "" {
		opts = append(opts, cmlclient.WithToken(token))
	} else {
		if username == "" || password == "" {
			fmt.Fprintln(os.Stderr, "provide -token/CML_TOKEN or -username/-password (CML_USERNAME/CML_PASSWORD)")
			os.Exit(2)
		}
		opts = append(opts, cmlclient.WithUsernamePassword(username, password))
	}

	c, err := cmlclient.New(address, opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "client init failed: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Force a system_information call so version is populated.
	if err := c.Ready(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "ready check failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(c.Version())
}
