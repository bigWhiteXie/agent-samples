package executor

import (
	"context"
	"io"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	ucb "github.com/cloudwego/eino/utils/callbacks"
)

func NewModelCallback(msgChan chan *schema.Message) callbacks.Handler {
	// 创建自定义回调处理器
	handler := &ucb.ModelCallbackHandler{
		OnEnd: func(ctx context.Context, runInfo *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
			msgChan <- output.Message
			if runInfo.Name == "reportLLM" {
				close(msgChan)
			}
			return ctx
		},
		OnEndWithStreamOutput: func(ctx context.Context, runInfo *callbacks.RunInfo, output *schema.StreamReader[*model.CallbackOutput]) context.Context {
			// 转换流为message
			converted := schema.StreamReaderWithConvert(output, func(m *model.CallbackOutput) (*schema.Message, error) {
				return m.Message, nil
			})

			go func() {
				for {
					msg, err := converted.Recv()
					if err == io.EOF {
						break
					}
					if err != nil {
						continue
					}
					msgChan <- msg
				}

				output.Close()
				converted.Close()
				if runInfo.Name == "reportLLM" {
					close(msgChan)
				}
			}()

			return ctx
		},
	}

	return ucb.NewHandlerHelper().ChatModel(handler).Handler()
}

func NewCloseCallback(msgChan chan *schema.Message) callbacks.Handler {
	return callbacks.NewHandlerBuilder().OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
		close(msgChan)
		return ctx
	}).OnErrorFn(func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
		close(msgChan)
		return ctx
	}).Build()
}
