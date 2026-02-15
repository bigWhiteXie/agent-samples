package workflow

import (
	"context"
	"fmt"
	"testing"

	"github.com/cloudwego/eino/compose"
)

func TestSample1(t *testing.T) {
	bidder1 := func(ctx context.Context, in float64) (float64, error) {
		return in + 1.0, nil
	}

	bidder2 := func(ctx context.Context, in float64) (float64, error) {
		return in + 2.0, nil
	}

	announcer := func(ctx context.Context, in any) (any, error) {
		fmt.Println("bidder1 had lodged his bid!")
		return nil, nil
	}

	wf := compose.NewWorkflow[float64, map[string]float64]()

	wf.AddLambdaNode("b1", compose.InvokableLambda(bidder1)).
		AddInput(compose.START)

	// just add a node to announce bidder1 had lodged his bid!
	// It should be executed strictly after bidder1, so we use `AddDependency("b1")`.
	// Note that `AddDependency()` will only form control relationship,
	// but not data passing relationship.
	wf.AddLambdaNode("announcer", compose.InvokableLambda(announcer)).
		AddDependency("b1")

	// add a branch just like adding branch in Graph.
	wf.AddBranch("b1", compose.NewGraphBranch(func(ctx context.Context, in float64) (string, error) {
		if in > 5.0 {
			return compose.END, nil
		}
		return "b2", nil
	}, map[string]bool{compose.END: true, "b2": true}))

	wf.AddLambdaNode("b2", compose.InvokableLambda(bidder2)).
		// b2 executes strictly after b1 (through branch dependency),
		// but does not rely on b1's output,
		// which means b2 depends on b1 conditionally,
		// but no data passing between them.
		AddInputWithOptions(compose.START, nil, compose.WithNoDirectDependency())

	wf.End().AddInput("b1", compose.ToField("bidder1")).
		AddInput("b2", compose.ToField("bidder2"))

	runner, err := wf.Compile(context.Background())
	if err != nil {
		fmt.Printf("workflow compile error: %v\n", err)
		return
	}

	result, err := runner.Invoke(context.Background(), 3.0)
	if err != nil {
		fmt.Printf("workflow run err: %v\n", err)
		return
	}

	fmt.Printf("%v\n", result)
}
