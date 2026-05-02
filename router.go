// Copyright 2026 Iyad
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bolt

import "sync"

// Param is a single URL path parameter, consisting of a key and a value.
type Param struct {
	// Key is the parameter name.
	Key string
	// Value is the parameter value extracted from the URL.
	Value string
}

// Params is a slice of path parameters extracted from the URL.
type Params []Param

// Get returns the value of the first Param with the given name, or "" if not found.
func (ps Params) Get(name string) string {
	for _, p := range ps {
		if p.Key == name {
			return p.Value
		}
	}

	return ""
}

// Router is an HTTP request multiplexer that routes requests
// using a radix tree for fast path matching.
type Router struct {
	trees       map[string]*node // method: rootNode
	middlewares []MiddlewareFunc
	maxParams   uint16
	paramsPool  sync.Pool
}

// New returns a new Router.
func New() *Router {
	return &Router{trees: make(map[string]*node)}
}

// Handle registers a handler for the given HTTP method and path pattern.
func (r *Router) Handle(method, path string, handler Handler) {
	if _, ok := r.trees[method]; !ok {
		r.trees[method] = &node{}
	}
	r.trees[method].addRoute(path, handler)

	if paramsCount := countParams(path); paramsCount > r.maxParams {
		r.maxParams = paramsCount
		r.paramsPool.New = func() any {
			ps := make(Params, 0, r.maxParams)
			return &ps
		}
	}
}

// GET registers a handler for GET requests.
func (r *Router) GET(path string, handler HandlerFunc) {
	r.Handle("GET", path, handler)
}

// POST registers a handler for POST requests.
func (r *Router) POST(path string, handler HandlerFunc) {
	r.Handle("POST", path, handler)
}

// PUT registers a handler for PUT requests.
func (r *Router) PUT(path string, handler HandlerFunc) {
	r.Handle("PUT", path, handler)
}

// DELETE registers a handler for DELETE requests.
func (r *Router) DELETE(path string, handler HandlerFunc) {
	r.Handle("DELETE", path, handler)
}

// PATCH registers a handler for PATCH requests.
func (r *Router) PATCH(path string, handler HandlerFunc) {
	r.Handle("PATCH", path, handler)
}

// Static registers a handler that serves static files from the given root directory.
func (r *Router) Static(prefix, root string) {
	r.GET(prefix+"/*filepath", fileServer(root))
}

// ServeHTTP dispatches the request to the matching route handler.
func (r *Router) ServeHTTP(w ResponseWriter, req *Request) {
	handler := HandlerFunc(func(w ResponseWriter, req *Request) {
		tree, ok := r.trees[req.Method]
		if !ok {
			w.WriteHeader(StatusMethodNotAllowed)
			w.Write([]byte(StatusText(StatusMethodNotAllowed)))
			return
		}

		hndlr, params, _ := tree.getValue(req.URL.Path, r.getParams)
		if params != nil {
			req.params = *params
		}

		if hndlr == nil {
			if params != nil {
				r.putParams(params)
			}
			w.WriteHeader(StatusNotFound)
			w.Write([]byte(StatusText(StatusNotFound)))
			return
		}

		hndlr.ServeHTTP(w, req)
		if params != nil {
			r.putParams(params)
		}
	})

	Chain(r.middlewares...)(handler).ServeHTTP(w, req)
}

// Use adds global middleware to the router.
func (r *Router) Use(middlewares ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middlewares...)
}

func (r *Router) getParams() *Params {
	ps, _ := r.paramsPool.Get().(*Params)
	*ps = (*ps)[:0]
	return ps
}

func (r *Router) putParams(ps *Params) {
	*ps = (*ps)[:0]
	r.paramsPool.Put(ps)
}
