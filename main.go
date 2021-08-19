package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vartanbeno/go-reddit/v2/reddit"
)

func main() {
	client, err := reddit.NewReadonlyClient()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	posts, _, err := client.Subreddit.TopPosts(context.Background(), "golang", &reddit.ListPostOptions{
		ListOptions: reddit.ListOptions{
			Limit: 5,
		},
		Time: "all",
	})
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("Received %d posts.\n", len(posts))
}
