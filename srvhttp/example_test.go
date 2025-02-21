package srvhttp_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/DoNewsCode/core"
	"github.com/DoNewsCode/core/srvhttp"
	"github.com/DoNewsCode/core/unierr"
	"github.com/gorilla/mux"
	"google.golang.org/grpc/codes"
)

func Example_modules() {
	c := core.New()
	defer c.Shutdown()

	c.AddModule(srvhttp.DocsModule{})
	c.AddModule(srvhttp.HealthCheckModule{})
	c.AddModule(srvhttp.MetricsModule{})
	c.AddModule(srvhttp.DebugModule{})

	router := mux.NewRouter()
	c.ApplyRouter(router)
	http.ListenAndServe(":8080", router)
}

func ExampleResponseEncoder_EncodeResponse() {
	handler := func(writer http.ResponseWriter, request *http.Request) {
		encoder := srvhttp.NewResponseEncoder(writer)
		encoder.EncodeResponse(struct {
			Foo string `json:"foo"`
		}{
			Foo: "bar",
		})
	}
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))

	// Output:
	// 200
	// application/json; charset=utf-8
	// {"foo":"bar"}
}

func ExampleResponseEncoder_EncodeError() {
	handler := func(writer http.ResponseWriter, request *http.Request) {
		encoder := srvhttp.NewResponseEncoder(writer)
		encoder.EncodeError(unierr.New(codes.NotFound, "foo is missing"))
	}
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(resp.StatusCode)
	fmt.Println(resp.Header.Get("Content-Type"))
	fmt.Println(string(body))

	// Output:
	// 404
	// application/json; charset=utf-8
	// {"code":5,"message":"foo is missing"}
}
