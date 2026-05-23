package main

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/slack-go/slack"
)

func TestSlackIntegrationOptIn(t *testing.T) {
	if os.Getenv("FEED2CLI_SLACK_INTEGRATION") != "1" {
		t.Skip("set FEED2CLI_SLACK_INTEGRATION=1 to run Slack integration test")
	}
	token := os.Getenv("CODEX_SLACK_BOT_TOKEN")
	channel := os.Getenv("CODEX_SLACK_CHANNEL")
	if token == "" {
		t.Skip("CODEX_SLACK_BOT_TOKEN is required")
	}
	api := slack.New(token)
	if channel == "" {
		var err error
		channel, err = slackIntegrationChannel(api)
		if err != nil {
			t.Fatalf("failed to prepare Slack integration channel: %v", err)
		}
	}

	if err := outputSlack([]*gofeed.Feed{{
		Items: []*gofeed.Item{{
			Title:       "feed2cli Slack integration test",
			Link:        "https://example.com/feed2cli-slack-integration",
			Description: "opt-in integration test",
		}},
	}}, api, channel, func(time.Duration) {}); err != nil {
		t.Fatalf("outputSlack returned error: %v", err)
	}
}

func slackIntegrationChannel(api *slack.Client) (string, error) {
	name := os.Getenv("CODEX_SLACK_TEST_CHANNEL_NAME")
	if name == "" {
		name = "feed2cli-codex-test"
	}
	name = strings.TrimPrefix(strings.TrimSpace(name), "#")
	if name == "" {
		name = "feed2cli-codex-test"
	}
	params := slackConversationListParams("")
	channels, _, err := api.GetConversations(&params)
	if err != nil {
		return "", err
	}
	for _, channel := range channels {
		if channel.Name == name || channel.NameNormalized == name {
			return channel.ID, nil
		}
	}
	isPrivate := os.Getenv("CODEX_SLACK_TEST_CHANNEL_PRIVATE") == "1" ||
		strings.EqualFold(os.Getenv("CODEX_SLACK_TEST_CHANNEL_PRIVATE"), "true")
	channel, err := api.CreateConversation(name, isPrivate)
	if err != nil {
		return "", err
	}
	return channel.ID, nil
}
