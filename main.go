package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/akrylysov/pogreb"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

const (
	limit               = 5
	messageQueueSize    = 100
	checkRedditDelay    = time.Minute * 1
	sendToTelegramDelay = time.Second * 10

	redditBaseURL = "https://www.reddit.com/"

	storeFile = "store.db"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	ticker := time.NewTicker(checkRedditDelay)

	messageQueue := make(chan reddit.Post, messageQueueSize)

	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		termChan := make(chan os.Signal, 1)
		signal.Notify(termChan, syscall.SIGTERM, syscall.SIGINT)
		<-termChan
		cancelFunc()
	}()

	redisClient, err := reddit.NewReadonlyClient()
	if err != nil {
		log.Fatalf("reddit client error: %v", err)
	}

	tgbot, err := tgbotapi.NewBotAPI(os.Getenv("TG_BOT_TOKEN"))
	if err != nil {
		log.Fatalf("telegram bot error: %v", err)
	}

	store, err := pogreb.Open(storeFile, nil)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}

	defer store.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go runTelegram(ctx, &wg, tgbot, messageQueue)

loop:
	for {
		select {
		case <-ctx.Done():
			fmt.Print("context is done, will shutdown gracefully")
			break loop
		default:
			//
		}

		select {
		case <-ticker.C:
			if err := runReddit(ctx, redisClient, store, messageQueue); err != nil {
				fmt.Printf("run error: %v", err)
			}
		case <-ctx.Done():
			fmt.Print("context is done, will shutdown gracefully")
			break loop
		}
	}

	wg.Wait()
}

func runReddit(ctx context.Context, client *reddit.Client, store *pogreb.DB, messageQueue chan reddit.Post) error {
	subreddit := os.Getenv("SUBREDDIT")

	posts, _, err := client.Subreddit.NewPosts(ctx, subreddit, &reddit.ListOptions{
		Limit: limit,
	})
	if err != nil {
		return fmt.Errorf("get new posts subreddit=%s: %v", subreddit, err)
	}

	lastID, err := store.Get([]byte(subreddit))
	if err != nil {
		return fmt.Errorf("get last id from store: %v", err)
	}

	if len(posts) == 0 {
		return nil
	}

	for _, post := range posts {
		if lastID != nil && post.FullID == string(lastID) {
			break
		}
		messageQueue <- *post
	}

	lastPostID := posts[0].FullID

	fmt.Printf("Received %d posts.\n", len(posts))
	fmt.Printf("LastID: %s\n", lastPostID)

	if err := store.Put([]byte(subreddit), []byte(lastPostID)); err != nil {
		return fmt.Errorf("put last id to store")
	}

	return nil
}

func runTelegram(ctx context.Context, wg *sync.WaitGroup, tgbot *tgbotapi.BotAPI, messageQueue chan reddit.Post) {
	defer wg.Done()

	chatIDRaw := os.Getenv("CHAT_ID")
	chatID, err := strconv.ParseInt(chatIDRaw, 10, 64)
	if err != nil {
		log.Fatalf("convert chat id to int64: %v", err)
	}

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		default:
			//
		}

		select {
		case <-ctx.Done():
			break loop
		case post := <-messageQueue:
			msg := fmt.Sprintf("%s%s", redditBaseURL, post.Permalink)
			tgMsg := tgbotapi.NewMessage(chatID, msg)
			if _, err := tgbot.Send(tgMsg); err != nil {
				fmt.Printf("telegram bot send message: %v\n", err)
			}
			time.Sleep(sendToTelegramDelay)
		}
	}
}
