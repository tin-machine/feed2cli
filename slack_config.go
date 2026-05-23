package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/slack-go/slack"
)

func slackTokenFromEnv() string {
	for _, name := range []string{"FEED2CLI_SLACK_TOKEN", "XOXB"} {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

func slackChannelFromOptions(channel string) string {
	channel = strings.TrimSpace(channel)
	if channel != "" {
		return channel
	}
	for _, name := range []string{"FEED2CLI_SLACK_CHANNEL", "SLACK_CHANNEL"} {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func prepareSlackDestination(api slackClient, channel string, validate bool) (string, error) {
	channel = strings.TrimSpace(channel)
	if channel == "" {
		return "", fmt.Errorf("Slack channel is empty")
	}
	if !validate {
		return channel, nil
	}
	if api == nil {
		return "", fmt.Errorf("slack client is nil")
	}
	if _, err := api.AuthTest(); err != nil {
		return "", fmt.Errorf("Slack auth.testに失敗しました: %w", err)
	}
	resolved, err := resolveSlackConversation(api, channel)
	if err != nil {
		return "", err
	}
	return resolved, nil
}

func resolveSlackConversation(api slackClient, channel string) (string, error) {
	channel = strings.TrimSpace(channel)
	if channel == "" {
		return "", fmt.Errorf("Slack channel is empty")
	}
	if strings.HasPrefix(channel, "#") {
		return findSlackConversationByName(api, strings.TrimPrefix(channel, "#"))
	}
	if looksLikeSlackConversationID(channel) {
		if _, err := api.GetConversationInfo(channel, false); err != nil {
			return "", fmt.Errorf("Slack channel %q を確認できません: %w", channel, err)
		}
		return channel, nil
	}
	if looksLikeSlackUserID(channel) {
		return channel, nil
	}
	return findSlackConversationByName(api, channel)
}

func findSlackConversationByName(api slackClient, name string) (string, error) {
	name = strings.TrimPrefix(strings.TrimSpace(name), "#")
	if name == "" {
		return "", fmt.Errorf("Slack channel name is empty")
	}
	cursor := ""
	for {
		params := slackConversationListParams(cursor)
		channels, nextCursor, err := api.GetConversations(&params)
		if err != nil {
			return "", fmt.Errorf("Slack channel name %q を解決できません: %w", name, err)
		}
		for _, channel := range channels {
			if channel.Name == name || channel.NameNormalized == name {
				return channel.ID, nil
			}
		}
		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}
	return "", fmt.Errorf("Slack channel %q が見つかりません。channel IDを指定するか、botが参加しているか確認してください", name)
}

func slackConversationListParams(cursor string) slack.GetConversationsParameters {
	return slack.GetConversationsParameters{
		Cursor:          cursor,
		ExcludeArchived: true,
		Limit:           200,
		Types:           []string{"public_channel", "private_channel"},
	}
}

func looksLikeSlackConversationID(value string) bool {
	if len(value) < 2 {
		return false
	}
	switch value[0] {
	case 'C', 'G', 'D':
		return true
	default:
		return false
	}
}

func looksLikeSlackUserID(value string) bool {
	return len(value) >= 2 && value[0] == 'U'
}
