package main

import (
	"fmt"
	"log"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

func main() {
	credentials := reddit.Credentials{ID: "id", Secret: "secret", Username: "username", Password: "password"}
	client, err := reddit.NewClient(credentials)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Print("test", client.ID)
}
