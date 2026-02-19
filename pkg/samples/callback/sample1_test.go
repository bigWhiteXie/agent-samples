package callback

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"testing"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func newChatModel() model.ToolCallingChatModel {
	cm, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		Model:   "qwen3",
		BaseURL: "http://127.0.0.1:11434/v1",
	})
	if err != nil {
		log.Fatal(err)
	}
	return cm
}

func TestGenerate(t *testing.T) {
	ctx := context.Background()
	cm := newChatModel()
	msg, err := cm.Generate(ctx, []*schema.Message{
		&schema.Message{
			Role:    schema.System,
			Content: "你是一名ai助手，请回答用户问题",
		},
		&schema.Message{
			Role:    schema.User,
			Content: "你会说英语吗",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Received message: %v", msg.Content)
}
func TestStream(t *testing.T) {
	ctx := context.Background()
	cm := newChatModel()
	reader, err := cm.Stream(ctx, []*schema.Message{
		&schema.Message{
			Role:    schema.System,
			Content: "你是一名ai助手，请回答用户问题",
		},
		&schema.Message{
			Role:    schema.User,
			Content: "你是谁",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	for {
		msg, err := reader.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
		}
		fmt.Print(msg.Content)
	}

}
