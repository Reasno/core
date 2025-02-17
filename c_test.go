package core

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/DoNewsCode/core/config"
	"github.com/DoNewsCode/core/contract"
	"github.com/DoNewsCode/core/di"
	"github.com/DoNewsCode/core/events"
	"github.com/DoNewsCode/core/otgorm"
	"github.com/DoNewsCode/core/srvgrpc"
	"github.com/DoNewsCode/core/srvhttp"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestC_Serve(t *testing.T) {
	var called int32
	c := New(
		WithInline("http.addr", ":19998"),
		WithInline("grpc.addr", ":19999"),
	)
	c.ProvideEssentials()
	c.AddModule(srvhttp.HealthCheckModule{})
	c.AddModule(srvgrpc.HealthCheckModule{})

	c.Invoke(func(dispatcher contract.Dispatcher) {
		dispatcher.Subscribe(events.Listen(OnHTTPServerStart, func(ctx context.Context, start interface{}) error {
			atomic.AddInt32(&called, 1)
			assert.Equal(t, "[::]:19998", start.(OnHTTPServerStartPayload).Listener.Addr().String())
			return nil
		}))
	})
	c.Invoke(func(dispatcher contract.Dispatcher) {
		dispatcher.Subscribe(events.Listen(OnHTTPServerShutdown, func(ctx context.Context, shutdown interface{}) error {
			atomic.AddInt32(&called, 1)
			assert.Equal(t, "[::]:19998", shutdown.(OnHTTPServerShutdownPayload).Listener.Addr().String())
			return nil
		}))
	})
	c.Invoke(func(dispatcher contract.Dispatcher) {
		dispatcher.Subscribe(events.Listen(OnGRPCServerStart, func(ctx context.Context, start interface{}) error {
			atomic.AddInt32(&called, 1)
			assert.Equal(t, "[::]:19999", start.(OnGRPCServerStartPayload).Listener.Addr().String())
			return nil
		}))
	})
	c.Invoke(func(dispatcher contract.Dispatcher) {
		dispatcher.Subscribe(events.Listen(OnGRPCServerShutdown, func(ctx context.Context, shutdown interface{}) error {
			atomic.AddInt32(&called, 1)
			assert.Equal(t, "[::]:19999", shutdown.(OnGRPCServerShutdownPayload).Listener.Addr().String())
			return nil
		}))
	})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	e := c.Serve(ctx)
	assert.NoError(t, e)
	assert.Equal(t, int32(4), atomic.LoadInt32(&called))
}

func TestC_ServeDisable(t *testing.T) {
	var called int32
	c := New(
		WithInline("http.disable", "true"),
		WithInline("grpc.disable", "true"),
		WithInline("cron.disable", "true"),
	)
	c.ProvideEssentials()
	c.AddModule(srvhttp.HealthCheckModule{})
	c.AddModule(srvgrpc.HealthCheckModule{})

	c.Invoke(func(dispatcher contract.Dispatcher) {
		dispatcher.Subscribe(events.Listen(OnHTTPServerStart, func(ctx context.Context, start interface{}) error {
			atomic.AddInt32(&called, 1)
			return nil
		}))
	})
	c.Invoke(func(dispatcher contract.Dispatcher) {
		dispatcher.Subscribe(events.Listen(OnHTTPServerShutdown, func(ctx context.Context, shutdown interface{}) error {
			atomic.AddInt32(&called, 1)
			return nil
		}))
	})
	c.Invoke(func(dispatcher contract.Dispatcher) {
		dispatcher.Subscribe(events.Listen(OnGRPCServerStart, func(ctx context.Context, start interface{}) error {
			atomic.AddInt32(&called, 1)
			return nil
		}))
	})
	c.Invoke(func(dispatcher contract.Dispatcher) {
		dispatcher.Subscribe(events.Listen(OnGRPCServerShutdown, func(ctx context.Context, shutdown interface{}) error {
			atomic.AddInt32(&called, 1)
			return nil
		}))
	})
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	e := c.Serve(ctx)
	assert.NoError(t, e)
	assert.Equal(t, int32(0), atomic.LoadInt32(&called))
}

func TestC_Default(t *testing.T) {
	c := New()
	c.ProvideEssentials()
	c.Provide(otgorm.Providers())
	c.AddModuleFunc(config.New)

	f, _ := ioutil.TempFile("./", "*")
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()

	rootCommand := &cobra.Command{}
	c.ApplyRootCommand(rootCommand)
	rootCommand.SetArgs([]string{"config", "init", "-o", f.Name()})
	rootCommand.Execute()

	output, _ := ioutil.ReadFile(f.Name())
	assert.Contains(t, string(output), "gorm:")
}

type m1 struct {
	di.Out
	A int
}

func (m m1) ModuleSentinel() {
	panic("implement me")
}

type m2 struct {
	di.Out
	A float32
}

func (m m2) ModuleSentinel() {
	panic("implement me")
}

func TestC_Provide(t *testing.T) {
	c := New()
	c.Provide(di.Deps{
		func() m1 { return m1{} },
		func() m2 { return m2{} },
	})
}

type a struct{}
type b struct{}

func mockConstructor(b b) (a, func(), error) {
	return a{}, func() {}, nil
}

func TestNew_missingDependencyErrorMessage(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			if strings.Contains(r.(error).Error(), "\"reflect\".makeFuncStub") {
				t.Error("should not contain reflection stub")
			}
			return
		}
		t.Error("test should panic")
	}()
	c := New()
	c.Provide(di.Deps{mockConstructor})
	c.Invoke(func(a a) error {
		return nil
	})
}
