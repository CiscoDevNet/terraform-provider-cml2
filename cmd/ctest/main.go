package main

import (
	"encoding/json"
	"fmt"

	"github.com/rschmied/terraform-provider-ciscocml/m/v2/pkg/cmlclient"
)

func main() {
	token := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJjb20uY2lzY28udmlybCIsImlhdCI6MTY1MTc1MDM0MywiZXhwIjoxNjUxODM2NzQzLCJzdWIiOiIwMDAwMDAwMC0wMDAwLTQwMDAtYTAwMC0wMDAwMDAwMDAwMDAifQ.xBGNWzKyOsUq22ALHwOYosMr7yVhNKFatdvFa5yOkA8"

	client := cmlclient.NewCMLClient("https://192.168.122.245", token, true)
	l, err := client.GetLab("52d5c824-e10c-450a-b9c5-b700bd3bc17a")
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Print(l)

	je, err := json.Marshal(l)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(je))
}
