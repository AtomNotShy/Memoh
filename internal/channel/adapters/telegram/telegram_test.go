package telegram

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func TestResolveTelegramSender(t *testing.T) {
	t.Parallel()

	externalID, displayName, attrs := resolveTelegramSender(nil)
	if externalID != "" || displayName != "" || len(attrs) != 0 {
		t.Fatalf("expected empty sender")
	}
	msg := &tgbotapi.Message{
		From: &tgbotapi.User{ID: 123, UserName: "alice"},
	}
	externalID, displayName, attrs = resolveTelegramSender(msg)
	if externalID != "123" || displayName != "alice" {
		t.Fatalf("unexpected sender: %s %s", externalID, displayName)
	}
	if attrs["user_id"] != "123" || attrs["username"] != "alice" {
		t.Fatalf("unexpected attrs: %#v", attrs)
	}
}

func TestIsTelegramBotMentioned(t *testing.T) {
	t.Parallel()

	t.Run("text mention", func(t *testing.T) {
		t.Parallel()
		msg := &tgbotapi.Message{
			Text: "hello @MemohBot",
		}
		if !isTelegramBotMentioned(msg, "memohbot") {
			t.Fatalf("expected bot mention from text")
		}
	})

	t.Run("entity text mention", func(t *testing.T) {
		t.Parallel()
		msg := &tgbotapi.Message{
			Entities: []tgbotapi.MessageEntity{
				{
					Type: "text_mention",
					User: &tgbotapi.User{IsBot: true},
				},
			},
		}
		if !isTelegramBotMentioned(msg, "") {
			t.Fatalf("expected bot mention from text_mention entity")
		}
	})

	t.Run("not mentioned", func(t *testing.T) {
		t.Parallel()
		msg := &tgbotapi.Message{
			Text: "hello everyone",
		}
		if isTelegramBotMentioned(msg, "memohbot") {
			t.Fatalf("expected no mention")
		}
	})
}
