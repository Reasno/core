package otetcd

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/DoNewsCode/core/config"
	"github.com/go-kit/kit/log"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3"
)

func TestTracing(t *testing.T) {
	if os.Getenv("ETCD_ADDR") == "" {
		t.Skip("Set env ETCD_ADDR to run TestTracing")
		return
	}
	addrs := strings.Split(os.Getenv("ETCD_ADDR"), ",")
	var interceptorCalled bool
	tracer := mocktracer.New()
	factory, cleanup := provideFactory(factoryIn{
		Logger: log.NewNopLogger(),
		Conf: config.MapAdapter{"etcd": map[string]Option{
			"default": {
				Endpoints: addrs,
			},
		}},
		Interceptor: func(name string, options *clientv3.Config) {
			interceptorCalled = true
			assert.Equal(t, "default", name)
		},
		Tracer: tracer,
	})
	defer cleanup()

	client, err := factory.Maker.Make("default")
	assert.NoError(t, err)
	span, ctx := opentracing.StartSpanFromContextWithTracer(context.Background(), tracer, "test")
	defer span.Finish()

	response, err := client.Get(ctx, "foo")
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, tracer.FinishedSpans())
	assert.True(t, interceptorCalled)

}
