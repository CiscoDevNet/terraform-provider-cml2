package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rschmied/terraform-provider-cml2/m/v2/internal/cmlclient"
)

func main() {
	host, found := os.LookupEnv("CML_HOST")
	if !found {
		fmt.Fprintln(os.Stderr, "CML_HOST env var not found!")
		return
	}
	token, found := os.LookupEnv("CML_TOKEN")
	if !found {
		fmt.Fprintln(os.Stderr, "CML_TOKEN env var not found!")
		return
	}
	labID, found := os.LookupEnv("CML_LABID")
	if !found {
		fmt.Fprintln(os.Stderr, "CML_LABID env var not found!")
		return
	}
	client := cmlclient.NewClient(host, token, true)
	l, err := client.GetLab(labID, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	je, err := json.Marshal(l)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	fmt.Println(string(je))
}
