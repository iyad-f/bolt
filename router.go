package bolt

// Param is a single URL path parameter, consisting of a key and a value.
type Param struct {
	Key   string
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

// ServeHTTP dispatches the request to the matching route handler.
func (r *Router) ServeHTTP(w ResponseWriter, req *Request) {
	tree, ok := r.trees[req.Method]
	if !ok {
		w.WriteHeader(StatusMethodNotAllowed)
		w.Write([]byte(StatusText(StatusMethodNotAllowed)))
		return
	}

	handler, _ := tree.search(req.URL.Path)
	if handler == nil {
		w.WriteHeader(StatusNotFound)
		w.Write([]byte(StatusText(StatusNotFound)))
		return
	}

	Chain(r.middlewares...)(handler).ServeHTTP(w, req)
}

// Use adds global middleware to the router.
func (r *Router) Use(middlewares ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, middlewares...)
}
