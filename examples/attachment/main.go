package main

import (
	"context"
	"fmt"
	"log"
	"os"

	mlievpush "github.com/mliev-sdk/push-go"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("usage: go run ./examples/attachment <recipient@example.com> <file>")
	}
	baseURL := os.Getenv("PUSH_BASE_URL")
	appID := os.Getenv("PUSH_APP_ID")
	secret := os.Getenv("PUSH_APP_SECRET")
	channelID := 0
	if _, err := fmt.Sscan(os.Getenv("PUSH_CHANNEL_ID"), &channelID); err != nil || channelID <= 0 {
		log.Fatal("set PUSH_CHANNEL_ID to an email channel ID")
	}
	if baseURL == "" || appID == "" || secret == "" {
		log.Fatal("set PUSH_BASE_URL, PUSH_APP_ID and PUSH_APP_SECRET")
	}

	attachment, err := mlievpush.NewEmailAttachmentFromFile(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	client := mlievpush.NewClient(baseURL, appID, secret)
	result, err := client.SendMessage(context.Background(), &mlievpush.SendMessageRequest{
		ChannelID:     channelID,
		SignatureName: os.Getenv("PUSH_SIGNATURE_NAME"),
		Receiver:      os.Args[1],
		Attachments:   []mlievpush.EmailAttachment{attachment},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("accepted task: %s\n", result.TaskID)
}
