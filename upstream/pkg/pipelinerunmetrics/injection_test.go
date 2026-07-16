package pipelinerunmetrics

import (
	"testing"

	_ "github.com/tektoncd/pipeline/pkg/client/injection/informers/pipeline/v1/pipelinerun/fake"
	"gotest.tools/v3/assert"
	rtesting "knative.dev/pkg/reconciler/testing"
)

func TestWithClient(t *testing.T) {
	ResetRecorder()
	ctx, _ := rtesting.SetupFakeContext(t)
	ctx = WithClient(ctx)

	rec := Get(ctx)
	assert.Assert(t, rec != nil)
}

func TestGetPanicsWithoutRecorder(t *testing.T) {
	defer func() {
		r := recover()
		assert.Assert(t, r != nil)
	}()
	ctx, _ := rtesting.SetupFakeContext(t)
	Get(ctx)
}

func TestWithInformer(t *testing.T) {
	ResetRecorder()
	ctx, cancel, _ := rtesting.SetupFakeContextWithCancel(t)
	ctx = WithClient(ctx)

	newCtx, informer := WithInformer(ctx)
	assert.Assert(t, informer != nil)
	assert.Assert(t, newCtx != nil)
	assert.Equal(t, informer.HasSynced(), true)

	stopCh := make(chan struct{})
	go informer.Run(stopCh)
	t.Cleanup(func() {
		close(stopCh)
	})
	cancel()
}
