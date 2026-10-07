package pipelinerunmetrics

import (
	"context"
	"testing"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/apis/pipelinesascode/keys"
	tektonv1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	fake "github.com/tektoncd/pipeline/pkg/client/clientset/versioned/fake"
	informers "github.com/tektoncd/pipeline/pkg/client/informers/externalversions"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"gotest.tools/v3/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

func TestCountRunningPRs(t *testing.T) {
	annotations := map[string]string{
		keys.GitProvider: "github",
		keys.EventType:   "pull_request",
		keys.Repository:  "pac-repo",
	}

	ctx := context.Background()
	var plrs []*tektonv1.PipelineRun
	pr := &tektonv1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   "pac-ns",
			Annotations: annotations,
		},
		Status: tektonv1.PipelineRunStatus{
			Status: duckv1.Status{Conditions: []apis.Condition{
				{
					Type:   apis.ConditionReady,
					Status: corev1.ConditionTrue,
					Reason: tektonv1.PipelineRunReasonRunning.String(),
				},
			}},
		},
	}

	numberOfRunningPRs := 10
	for i := 0; i < numberOfRunningPRs; i++ {
		plrs = append(plrs, pr)
	}

	ResetRecorder()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)
	m, err := NewRecorder()
	assert.NilError(t, err)

	_, err = m.meter.RegisterCallback(func(_ context.Context, o metric.Observer) error {
		return m.ObserveRunningPRsMetrics(o, plrs)
	}, m.runningPRCount)
	assert.NilError(t, err)

	var rm metricdata.ResourceMetrics
	err = reader.Collect(ctx, &rm)
	assert.NilError(t, err, "error collecting metrics")

	assert.Equal(t, len(rm.ScopeMetrics), 1)
	assert.Equal(t, len(rm.ScopeMetrics[0].Metrics), 1)
	assert.Equal(t, rm.ScopeMetrics[0].Metrics[0].Name, "pipelines_as_code_running_pipelineruns_count")
	count, ok := rm.ScopeMetrics[0].Metrics[0].Data.(metricdata.Gauge[int64])
	assert.Assert(t, ok)
	assert.Equal(t, count.DataPoints[0].Value, int64(numberOfRunningPRs))
}

func TestRecorderMetrics(t *testing.T) {
	ResetRecorder()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)
	m, err := NewRecorder()
	assert.NilError(t, err)

	ctx := context.Background()
	assert.NilError(t, m.Count(ctx, "github", "pull_request", "ns", "repo"))
	assert.NilError(t, m.CountPRDuration(ctx, "ns", "repo", "succeeded", "", 10))
	assert.NilError(t, m.ReportGitProviderAPIUsage("github", "pull_request", "ns", "repo"))

	var rm metricdata.ResourceMetrics
	err = reader.Collect(ctx, &rm)
	assert.NilError(t, err, "error collecting metrics")

	names := map[string]bool{}
	for _, sm := range rm.ScopeMetrics {
		for _, met := range sm.Metrics {
			names[met.Name] = true
		}
	}
	assert.Assert(t, names["pipelines_as_code_pipelinerun_count"])
	assert.Assert(t, names["pipelines_as_code_pipelinerun_duration_seconds_sum"])
	assert.Assert(t, names["pipelines_as_code_git_provider_api_request_count"])
}

func TestRecorderNotInitialized(t *testing.T) {
	ResetRecorder()
	r := &Recorder{}
	ctx := context.Background()

	assert.Assert(t, r.Count(ctx, "github", "pull_request", "ns", "repo") != nil)
	assert.Assert(t, r.CountPRDuration(ctx, "ns", "repo", "succeeded", "", 10) != nil)
	assert.Assert(t, r.ReportGitProviderAPIUsage("github", "pull_request", "ns", "repo") != nil)
}

func TestObserveRunningPRsMetricsEmpty(t *testing.T) {
	ResetRecorder()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)
	m, err := NewRecorder()
	assert.NilError(t, err)

	err = m.ObserveRunningPRsMetrics(nil, nil)
	assert.NilError(t, err)
}

func TestReportRunningPipelineRuns(t *testing.T) {
	ResetRecorder()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)
	m, err := NewRecorder()
	assert.NilError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	fakeClient := fake.NewSimpleClientset()
	factory := informers.NewSharedInformerFactory(fakeClient, 0)
	lister := factory.Tekton().V1().PipelineRuns().Lister()

	done := make(chan struct{})
	go func() {
		m.ReportRunningPipelineRuns(ctx, lister)
		close(done)
	}()
	cancel()
	<-done
}
