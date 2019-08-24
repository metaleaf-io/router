/*
   Copyright 2019 Metaleaf.io

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/
package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRouter(t *testing.T) {
	r := NewRouter()
	if r == nil {
		t.Error("Response must not be nil")
	}
}

func TestRouter_AddRoute(t *testing.T) {
	r := NewRouter()
	r.AddRoute("GET", "/path/{foo}/{bar}", func(writer http.ResponseWriter, request *Request) {
	})

	if r.routes == nil {
		t.Error("Routes must not be nil")
	}

	if len(r.routes) != 1 {
		t.Error("Routes must contain one route")
	}

	s := r.routes[0]

	if s.handler == nil {
		t.Error("Route handler must not be nil")
	}

	actual := s.path.String()
	expected := "^/path/(?P<foo>.+?)/(?P<bar>.+?)$"
	if actual != expected {
		t.Errorf("Path regex failed. Expected: \"%s\" Actual: \"%s\"", expected, actual)
	}

	if s.verb != "GET" {
		t.Error("Route verb must be GET")
	}
}

func TestRouter_ServeHTTP(t *testing.T) {
	// Create the router
	router := NewRouter()
	router.AddRoute("GET", "/path/{foo}/{bar}", func(writer http.ResponseWriter, request *Request) {
		if len(request.Params) != 4 {
			// Test for the params map.
			t.Errorf("Param failed. Expected \"4\" Actual: \"%d\"", len(request.Params))
		}

		p := request.Params["foo"]
		if p != "fuz" {
			t.Errorf("Param foo failed. Expected:fuz Actual:%s", p)
		}

		p = request.Params["bar"]
		if p != "baz" {
			t.Errorf("Param bar failed. Expected:baz Actual:%s", p)
		}

		p = request.Params["aaa"]
		if p != "bbb" {
			t.Errorf("Param aaa failed. Expected:baz Actual:%s", p)
		}

		p = request.Params["ccc"]
		if p != "ddd" {
			t.Errorf("Param ccc failed. Expected:baz Actual:%s", p)
		}

		writer.WriteHeader(200)
	})

	router.AddRoute("GET", "/", func(writer http.ResponseWriter, request *Request) {
		writer.WriteHeader(200)
	})

	server := httptest.NewServer(router)
	defer server.Close()

	request := server.URL + "/path/fuz/baz?aaa=bbb&ccc=ddd"
	res, err := http.Get(request)
	if err != nil {
		t.Errorf("GET failed with %v", err)
	}

	if res.StatusCode != 200 {
		t.Errorf("Response code failed. Expected:200 Actual:%d", res.StatusCode)
	}

	request = server.URL + "/"
	res, err = http.Get(request)
	if err != nil {
		t.Errorf("Get / failed with %v", err)
	}
	if res.StatusCode != 200 {
		t.Error("GET / expected 200 actual", res.StatusCode)
	}
}

// Examples that describe how to implement route handlers that accept query
// parameters and path variables.
func Example() {
	// Sample search handler.
	search := func(writer http.ResponseWriter, request *Request) {
		fmt.Printf("Search for: \"%s\"\n", request.Params["s"])
		writer.WriteHeader(200)
	}

	// Sample get resource handler.
	getBookByIsbn := func(writer http.ResponseWriter, request *Request) {
		fmt.Printf("Get book with ISBN = %s\n", request.Params["isbn"])
		writer.WriteHeader(200)
	}

	// build the router.
	router := NewRouter().
		AddRoute("GET", "/search", search).
		AddRoute("GET", "/book/{isbn}", getBookByIsbn)

	// Start the server.
	server := httptest.NewServer(router)
	defer server.Close()

	http.Get(server.URL + "/search?s=stuff+to+search+for")
	http.Get(server.URL + "/book/978-0316371247")

	// OUTPUT: Search for: "stuff to search for"
	// Get book with ISBN = 978-0316371247
}
