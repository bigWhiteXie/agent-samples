package checkpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/cloudwego/eino/compose"
)

type MyStore struct {
	store map[string][]byte
}

func NewStore() *MyStore {
	return &MyStore{
		store: make(map[string][]byte),
	}
}

func (s *MyStore) Get(ctx context.Context, key string) (value []byte, existed bool, err error) {
	val, ok := s.store[key]
	return val, ok, nil
}
func (s *MyStore) Set(ctx context.Context, key string, value []byte) (err error) {
	s.store[key] = value
	return nil
}
func TestS1(t *testing.T) {
	g := compose.NewGraph[string, string]()
	err := g.AddLambdaNode("node1", compose.InvokableLambda(func(ctx context.Context, input string) (output string, err error) { return input + " node1", nil }))
	if err != nil { /* error handle */
	}
	err = g.AddLambdaNode("node2", compose.InvokableLambda(func(ctx context.Context, input string) (output string, err error) { return input + " node2", nil }))
	if err != nil { /* error handle */
	}

	/** other graph composed code
	  xxx
	*/
	g.AddEdge(compose.START, "node1")
	g.AddEdge("node1", "node2")
	g.AddEdge("node2", compose.END)

	r, err := g.Compile(context.TODO(), compose.WithInterruptAfterNodes([]string{"node1"}), compose.WithInterruptBeforeNodes([]string{"node2"}))
	if err != nil {
		t.Errorf("%s", err)
	}
	_, err = r.Invoke(context.TODO(), "hello")
	if err != nil {
		if info, ok := compose.ExtractInterruptInfo(err); ok {
			infoStr, _ := json.Marshal(info)
			t.Logf("%s", infoStr)
		} else {
			t.Errorf("%s", err)
		}
	}
}

func TestS2(t *testing.T) {
	g := compose.NewGraph[string, string]()
	err := g.AddLambdaNode("node1", compose.InvokableLambda(func(ctx context.Context, input string) (output string, err error) {
		return input + " node1", compose.Interrupt(ctx, "需要用户补充信息")
	}))
	if err != nil { /* error handle */
	}
	err = g.AddLambdaNode("node2", compose.InvokableLambda(func(ctx context.Context, input string) (output string, err error) { return input + " node2", nil }))
	if err != nil { /* error handle */
	}

	/** other graph composed code
	  xxx
	*/
	g.AddEdge(compose.START, "node1")
	g.AddEdge("node1", "node2")
	g.AddEdge("node2", compose.END)

	r, err := g.Compile(context.TODO())
	if err != nil {
		t.Errorf("%s", err)
	}
	_, err = r.Invoke(context.TODO(), "hello")
	if err != nil {
		if info, ok := compose.ExtractInterruptInfo(err); ok {
			infoStr, _ := json.Marshal(info)
			t.Logf("%s", infoStr)
		} else {
			t.Errorf("%s", err)
		}
	}
}

func TestS3(t *testing.T) {
	g := compose.NewGraph[string, string]()
	err := g.AddLambdaNode("node1", compose.InvokableLambda(func(ctx context.Context, input string) (output string, err error) {
		if strings.Contains(input, "info") {
			return "pass", nil
		}
		return "", compose.Interrupt(ctx, "需要用户补充信息")
	}))
	if err != nil { /* error handle */
	}
	err = g.AddLambdaNode("node2", compose.InvokableLambda(func(ctx context.Context, input string) (output string, err error) {
		isResume, hasData, data := compose.GetResumeContext[string](ctx)
		if isResume && hasData && strings.Contains(data, "ok") {
			return "node2 pass", nil
		}

		if strings.Contains(input, "ok") {
			return "node2 pass", nil
		}

		return "", compose.StatefulInterrupt(ctx, "需要用户补充node2信息", input)
	}))
	if err != nil { /* error handle */
	}

	/** other graph composed code
	  xxx
	*/
	g.AddEdge(compose.START, "node1")
	g.AddEdge("node1", "node2")
	g.AddEdge("node2", compose.END)

	store := NewStore()
	r, err := g.Compile(context.TODO(), compose.WithCheckPointStore(store))
	if err != nil {
		t.Errorf("%s", err)
	}
	_, err = r.Invoke(context.TODO(), "info", compose.WithCheckPointID("cp1"))
	var interruptID string
	if err != nil {
		if info, ok := compose.ExtractInterruptInfo(err); ok {
			interruptID = info.InterruptContexts[0].ID
			// infoStr, _ := json.Marshal(info)
			// t.Logf("%s", infoStr)
		} else {
			t.Errorf("%s", err)
		}
	}
	for k, v := range store.store {
		fmt.Printf("%s->%s\n", k, v)
	}

	// 或携带恢复数据（常用）
	resCtx := compose.ResumeWithData(context.Background(), interruptID, "ok")
	output, err := r.Invoke(resCtx, "", compose.WithCheckPointID("cp1"))
	if err != nil {
		t.Error(err)
	}
	t.Logf("%s", output)
}
