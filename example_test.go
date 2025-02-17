package core_test

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/di"
	"github.com/gorilla/mux"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/basicflag"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

func ExampleC_AddModuleFunc() {
	type Foo struct{}
	c := core.New()
	c.AddModuleFunc(func() (Foo, func(), error) {
		return Foo{}, func() {}, nil
	})
	fmt.Printf("%T\n", c.Modules()...)
	// Output:
	// core_test.Foo
}

func ExampleC_AddModule() {
	type Foo struct{}
	c := core.New()
	c.AddModule(Foo{})
	fmt.Printf("%T\n", c.Modules()...)
	// Output:
	// core_test.Foo
}

func ExampleC_Provide() {
	type Foo struct {
		Value string
	}
	type Bar struct {
		foo Foo
	}
	c := core.New()
	c.Provide(di.Deps{
		func() (foo Foo, cleanup func(), err error) {
			return Foo{
				Value: "test",
			}, func() {}, nil
		},
		func(foo Foo) Bar {
			return Bar{foo: foo}
		},
	})
	c.Invoke(func(bar Bar) {
		fmt.Println(bar.foo.Value)
	})
	// Output:
	// test
}

func Example_minimal() {

	// Spin up a real server
	c := core.Default(core.WithInline("log.level", "none"))
	c.AddModule(core.HttpFunc(func(router *mux.Router) {
		router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte("hello world"))
		})
	}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go c.Serve(ctx)

	// Giver server sometime to be ready.
	time.Sleep(time.Second)

	// Let's try if the server works.
	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	bytes, _ := ioutil.ReadAll(resp.Body)
	cancel()

	fmt.Println(string(bytes))
	// Output:
	// hello world
}

func ExampleC_stack() {

	fs := flag.NewFlagSet("example", flag.ContinueOnError)
	fs.String("log.level", "error", "the log level")
	// Spin up a real server
	c := core.New(
		core.WithConfigStack(basicflag.Provider(fs, "."), nil),
		core.WithConfigStack(env.Provider("APP_", ".", func(s string) string {
			return strings.ToLower(strings.Replace(s, "APP_", "", 1))
		}), nil),
		core.WithConfigStack(file.Provider("./config/testdata/mock.json"), json.Parser()),
	)
	c.AddModule(core.HttpFunc(func(router *mux.Router) {
		router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
			writer.Write([]byte("hello world"))
		})
	}))
	c.Serve(context.Background())
}
