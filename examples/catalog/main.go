// This example lists configuration by default. Sending requires an explicit -send flag.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	mlievpush "github.com/mliev-sdk/push-go"
)

func main() {
	channelID := flag.Int("channel", 0, "channel ID selected from the list")
	signature := flag.String("signature", "", "alias selected from signature_names")
	paramsJSON := flag.String("params", "{}", "template variable values as a JSON object of strings")
	receiver := flag.String("receiver", "", "recipient for sending")
	send := flag.Bool("send", false, "send after validating the selected configuration")
	flag.Parse()
	baseURL, appID, secret := os.Getenv("PUSH_BASE_URL"), os.Getenv("PUSH_APP_ID"), os.Getenv("PUSH_APP_SECRET")
	if baseURL == "" || appID == "" || secret == "" {
		log.Fatal("set PUSH_BASE_URL, PUSH_APP_ID and PUSH_APP_SECRET")
	}
	client := mlievpush.NewClient(baseURL, appID, secret)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	page, err := client.ListChannels(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	printJSON(page)
	if *channelID == 0 {
		if *send {
			log.Fatal("select a -channel before sending")
		}
		return
	}
	detail, err := client.GetChannel(ctx, *channelID)
	if err != nil {
		log.Fatal(err)
	}
	printJSON(detail)
	if !*send {
		return
	}
	if detail.Readiness.State == mlievpush.ChannelReadinessBlocked || detail.Template == nil || detail.Template.Variables == nil {
		log.Fatalf("channel unavailable: %v", detail.Readiness.BlockerCodes)
	}
	var params map[string]string
	if err := json.Unmarshal([]byte(*paramsJSON), &params); err != nil {
		log.Fatal(err)
	}
	for _, variable := range detail.Template.Variables {
		if _, exists := params[variable]; !exists {
			log.Fatalf("missing template variable: %s", variable)
		}
	}
	if detail.SignatureRequired || *signature != "" {
		found := false
		for _, alias := range detail.SignatureNames {
			found = found || alias == *signature
		}
		if !found {
			log.Fatal("choose -signature from signature_names")
		}
	}
	if *receiver == "" {
		log.Fatal("provide -receiver for sending")
	}
	result, err := client.SendMessage(ctx, &mlievpush.SendMessageRequest{
		ChannelID: detail.ID, Receiver: *receiver, SignatureName: *signature, TemplateParams: params,
	})
	if err != nil {
		log.Fatal(err)
	}
	printJSON(result)
}

func printJSON(value interface{}) {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(raw))
}
