package utils

import (
	cntx "context"
	"strings"

	gogpt "github.com/sashabaranov/go-gpt3"
	tele "gopkg.in/telebot.v3"
)

func ChatGPT(context tele.Context) error {
	if context.Message().ReplyTo == nil || (context.Message().ReplyTo.Sender.ID != Bot.Me.ID && strings.HasPrefix(context.Message().Text, "@"+Bot.Me.Username)) || context.Message().Text[:1] == "/" {
		return nil
	}
	c := gogpt.NewClient(Config.OpenAIKey)
	ctx := cntx.Background()

	req := gogpt.ChatCompletionRequest{
		Model:    gogpt.GPT3Dot5Turbo,
		Messages: []gogpt.ChatCompletionMessage{{Role: "user", Content: context.Message().Text}},
	}
	resp, err := c.CreateChatCompletion(ctx, req)
	if err != nil {
		return err
	}
	return context.Reply(resp.Choices[0].Message.Content)
}
