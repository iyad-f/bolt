package bolt

import "testing"

func dummyHandler(w ResponseWriter, r *Request) {}

func TestTree(t *testing.T) {
	t.Run("static routes", func(t *testing.T) {
		n := &node{}
		n.addRoute("/users", HandlerFunc(dummyHandler))
		n.addRoute("/users/create", HandlerFunc(dummyHandler))
		n.addRoute("/posts", HandlerFunc(dummyHandler))

		handler, _ := n.search("/users")
		if handler == nil {
			t.Error("/users: expected handler, got nil")
		}

		handler, _ = n.search("/users/create")
		if handler == nil {
			t.Error("/users/create: expected handler, got nil")
		}

		handler, _ = n.search("/posts")
		if handler == nil {
			t.Error("/posts: expected handler, got nil")
		}
	})

	t.Run("param routes", func(t *testing.T) {
		n := &node{}
		n.addRoute("/users/:id", HandlerFunc(dummyHandler))

		handler, params := n.search("/users/42")
		if handler == nil {
			t.Fatal("/users/42: expected handler, got nil")
		}
		if params.Get("id") != "42" {
			t.Errorf("param id = %q, want %q", params.Get("id"), "42")
		}
	})

	t.Run("param with trailing path", func(t *testing.T) {
		n := &node{}
		n.addRoute("/users/:id/posts", HandlerFunc(dummyHandler))

		handler, params := n.search("/users/42/posts")
		if handler == nil {
			t.Fatal("/users/42/posts: expected handler, got nil")
		}
		if params.Get("id") != "42" {
			t.Errorf("param id = %q, want %q", params.Get("id"), "42")
		}
	})

	t.Run("multiple params", func(t *testing.T) {
		n := &node{}
		n.addRoute("/users/:userId/posts/:postId", HandlerFunc(dummyHandler))

		handler, params := n.search("/users/42/posts/99")
		if handler == nil {
			t.Fatal("/users/42/posts/99: expected handler, got nil")
		}
		if params.Get("userId") != "42" {
			t.Errorf("param userId = %q, want %q", params.Get("userId"), "42")
		}
		if params.Get("postId") != "99" {
			t.Errorf("param postId = %q, want %q", params.Get("postId"), "99")
		}
	})

	t.Run("catch-all", func(t *testing.T) {
		n := &node{}
		n.addRoute("/static/*filepath", HandlerFunc(dummyHandler))

		handler, params := n.search("/static/css/style.css")
		if handler == nil {
			t.Fatal("/static/css/style.css: expected handler, got nil")
		}
		if params.Get("filepath") != "css/style.css" {
			t.Errorf("param filepath = %q, want %q", params.Get("filepath"), "css/style.css")
		}
	})

	t.Run("not found", func(t *testing.T) {
		n := &node{}
		n.addRoute("/users", HandlerFunc(dummyHandler))

		handler, _ := n.search("/posts")
		if handler != nil {
			t.Error("/posts: expected nil, got handler")
		}
	})

	t.Run("static over param priority", func(t *testing.T) {
		n := &node{}
		n.addRoute("/users/create", HandlerFunc(func(w ResponseWriter, r *Request) {}))
		n.addRoute("/users/:id", HandlerFunc(func(w ResponseWriter, r *Request) {}))

		handler, params := n.search("/users/create")
		if handler == nil {
			t.Fatal("/users/create: expected handler, got nil")
		}
		if params != nil && params.Get("id") != "" {
			t.Error("/users/create: should match static, not param")
		}

		handler, params = n.search("/users/42")
		if handler == nil {
			t.Fatal("/users/42: expected handler, got nil")
		}
		if params.Get("id") != "42" {
			t.Errorf("param id = %q, want %q", params.Get("id"), "42")
		}
	})

	t.Run("shared prefix splitting", func(t *testing.T) {
		n := &node{}
		n.addRoute("/users", HandlerFunc(dummyHandler))
		n.addRoute("/uploads", HandlerFunc(dummyHandler))

		handler, _ := n.search("/users")
		if handler == nil {
			t.Error("/users: expected handler, got nil")
		}
		handler, _ = n.search("/uploads")
		if handler == nil {
			t.Error("/uploads: expected handler, got nil")
		}
	})
}
