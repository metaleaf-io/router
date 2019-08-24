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
	"github.com/metaleaf-io/log"
	"net/http"
	"regexp"
	"strings"
)

// Enhances the regular http.Request structure by adding the parameters
type Request struct {
	// The original request structure.
	*http.Request

	// Parameters are a hash of key/value strings. These are extracted from
	// the URL path and the query that appears after a question mark "?" in
	// the path.
	Params map[string]string
}

// Prototype for the handler function.
type Handler func(http.ResponseWriter, *Request)

// Stores routes added by the application.
type Router struct {
	routes []route
}

// Describes a single route as a combination of HTTP VERB, regular expression
// path matcher, and the handler function.
type route struct {
	verb    string
	path    *regexp.Regexp
	handler Handler
}

// Some globals to make life easier.
var (
	paramRE = regexp.MustCompile("{(.+?)}")
)

// NewRouter initializes a new HTTP request httprouter.
func NewRouter() *Router {
	return new(Router)
}

// Adds a new route with a handler function. The router structure is also
// returned to allow chaining.
func (router *Router) AddRoute(verb string, path string, handler Handler) *Router {
	log.Info("Adding route", log.String("verb", verb), log.String("path", path))

	// Converts params in the path from "{param}" to a non-greedy regex named
	// match, "(?P<param>.+?)"
	if path != "/" {
		path = strings.TrimRight(path, "/")
		submatches := paramRE.FindAllString(path, -1)
		for _, s := range submatches {
			path = strings.Replace(path, s, "(?P<"+strings.Trim(s, "{}")+">.+?)", 1)
		}
		path = "^" + path + "$"
	}

	// Compile the path regex
	re, err := regexp.Compile(path)
	if err != nil {
		log.Error("Invalid path regex", log.Err("error", err))
	}

	// Adds the route if no errors occurred the regex compiler.
	var r route
	r.handler = handler
	r.path = re
	r.verb = verb

	router.routes = append(router.routes, r)
	return router
}

// Default global request handler that matches the incoming request with a
// registered handler.
func (router *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	for _, r := range router.routes {
		if request.Method == r.verb && r.path.MatchString(request.URL.Path) {
			m := matches(r.path, request.URL.Path)
			for k, v := range request.URL.Query() {
				m[k] = strings.Join(v, "; ")
			}

			httpRequest := new(Request)
			httpRequest.Request = request
			httpRequest.Params = m

			r.handler(writer, httpRequest)
			return
		}
	}
	log.Warn("Path not found", log.String("path", request.URL.Path))
	writer.WriteHeader(404)
}

// Helper that applies the path regex to the incoming path to parse param
// values from it.
func matches(re *regexp.Regexp, s string) map[string]string {
	submatches := re.FindStringSubmatch(s)
	matches := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i > 0 && name != "" {
			matches[name] = submatches[i]
		}
	}
	return matches
}
