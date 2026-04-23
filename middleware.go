package bolt

// MiddlewareFunc is a function that wraps a Handler to add behavior before or after it.
type MiddlewareFunc func(Handler) Handler

// Chain composes multiple middlewares into one. Middlewares are applied in the order given,
// so Chain(A, B, C) applied to handler H produces A(B(C(H))).
func Chain(middlewares ...MiddlewareFunc) MiddlewareFunc {
	return func(final Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}
