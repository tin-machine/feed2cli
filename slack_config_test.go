package main

import (
	"errors"
	"testing"

	"github.com/slack-go/slack"
)

type fakeSlackClient struct {
	fakeSlackPoster
	authErr          error
	infoErr          error
	conversationsErr error
	conversations    []slack.Channel
	infoCalls        []string
}

func (f *fakeSlackClient) AuthTest() (*slack.AuthTestResponse, error) {
	if f.authErr != nil {
		return nil, f.authErr
	}
	return &slack.AuthTestResponse{Team: "test"}, nil
}

func (f *fakeSlackClient) GetConversationInfo(channelID string, includeLocale bool) (*slack.Channel, error) {
	f.infoCalls = append(f.infoCalls, channelID)
	if f.infoErr != nil {
		return nil, f.infoErr
	}
	return &slack.Channel{GroupConversation: slack.GroupConversation{Conversation: slack.Conversation{ID: channelID}}}, nil
}

func (f *fakeSlackClient) GetConversations(params *slack.GetConversationsParameters) ([]slack.Channel, string, error) {
	if f.conversationsErr != nil {
		return nil, "", f.conversationsErr
	}
	return f.conversations, "", nil
}

func TestPrepareSlackDestinationValidatesConversationID(t *testing.T) {
	client := &fakeSlackClient{}

	got, err := prepareSlackDestination(client, "C123", true)
	if err != nil {
		t.Fatalf("prepareSlackDestination returned error: %v", err)
	}
	if got != "C123" {
		t.Fatalf("channel = %q, want C123", got)
	}
	if len(client.infoCalls) != 1 || client.infoCalls[0] != "C123" {
		t.Fatalf("infoCalls = %#v", client.infoCalls)
	}
}

func TestPrepareSlackDestinationResolvesChannelName(t *testing.T) {
	client := &fakeSlackClient{conversations: []slack.Channel{
		{GroupConversation: slack.GroupConversation{
			Conversation: slack.Conversation{ID: "C999", NameNormalized: "feed2cli-test"},
			Name:         "feed2cli-test",
		}},
	}}

	got, err := prepareSlackDestination(client, "#feed2cli-test", true)
	if err != nil {
		t.Fatalf("prepareSlackDestination returned error: %v", err)
	}
	if got != "C999" {
		t.Fatalf("channel = %q, want C999", got)
	}
}

func TestPrepareSlackDestinationReturnsValidationErrors(t *testing.T) {
	_, err := prepareSlackDestination(&fakeSlackClient{authErr: errors.New("invalid_auth")}, "C123", true)
	if err == nil {
		t.Fatal("expected auth error, got nil")
	}

	_, err = prepareSlackDestination(&fakeSlackClient{infoErr: errors.New("channel_not_found")}, "C123", true)
	if err == nil {
		t.Fatal("expected channel error, got nil")
	}
}

func TestPrepareSlackDestinationCanSkipValidation(t *testing.T) {
	got, err := prepareSlackDestination(nil, "#maybe-name", false)
	if err != nil {
		t.Fatalf("prepareSlackDestination returned error: %v", err)
	}
	if got != "#maybe-name" {
		t.Fatalf("channel = %q", got)
	}
}
