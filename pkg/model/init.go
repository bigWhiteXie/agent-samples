package model

import (
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

func OpenAI(ctx context.Context) (model.ToolCallingChatModel, error) {
	return openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: "https://api.siliconflow.cn/v1",
		Model:   "deepseek-ai/DeepSeek-V3.2",                           // 使用的模型版本
		APIKey:  "sk-xkodktqbtxywdsjgalwaqsjrqcghddoihycnqvjkkhwcynho", // OpenAI API 密钥
	})
}
