package theatre

import (
	"net/http"
	"strings"
)

// Route represents a route with a pattern, handler, and HTTP method
type Route struct {
	Pattern string
	Handler http.HandlerFunc
	Method  string
}

// CustomRouter holds the registered routes
type CustomRouter struct {
	routes []Route
}

// NewRouter creates a new instance of CustomRouter
func NewRouter() *CustomRouter {
	return &CustomRouter{
		routes: []Route{},
	}
}

// Handle registers a new route with the given pattern, handler and HTTP method
func (r *CustomRouter) Handle(pattern, method string, handler http.HandlerFunc) {
	r.routes = append(r.routes, Route{
		Pattern: pattern,
		Handler: handler,
		Method:  method,
	})
}

// GET registers a new GET route
func (r *CustomRouter) GET(pattern string, handler http.HandlerFunc) {
	r.Handle(pattern, http.MethodGet, handler)
}

// POST registers a new POST route
func (r *CustomRouter) POST(pattern string, handler http.HandlerFunc) {
	r.Handle(pattern, http.MethodPost, handler)
}

// PUT registers a new PUT route
func (r *CustomRouter) PUT(pattern string, handler http.HandlerFunc) {
	r.Handle(pattern, http.MethodPut, handler)
}

// DELETE registers a new DELETE route
func (r *CustomRouter) DELETE(pattern string, handler http.HandlerFunc) {
	r.Handle(pattern, http.MethodDelete, handler)
}

// OPTIONS registers a new OPTIONS route
func (r *CustomRouter) OPTIONS(pattern string, handler http.HandlerFunc) {
	r.Handle(pattern, http.MethodOptions, handler)
}

// ServeHTTP implements the http.Handler interface
func (r *CustomRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.routes {
		// First check if HTTP method matches (unless it's OPTIONS, which we'll handle specially)
		if req.Method != route.Method && req.Method != http.MethodOptions {
			continue
		}

		// Check if the route pattern matches
		params, matched := matchPattern(route.Pattern, req.URL.Path)
		if !matched {
			continue
		}

		// Store route parameters in the request context
		ctx := req.Context()
		for key, value := range params {
			ctx = contextWithParam(ctx, key, value)
		}

		// If it's an OPTIONS request and the path matches, handle CORS
		if req.Method == http.MethodOptions {
			handleCORS(w)
			return
		}

		// Apply CORS middleware for non-OPTIONS requests
		CORSMiddleware(route.Handler)(w, req.WithContext(ctx))
		return
	}

	// No route matched
	http.NotFound(w, req)
}

// Helper function to handle CORS for OPTIONS requests
func handleCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
	w.WriteHeader(http.StatusNoContent)
}

// Helper function to split a path by "/"
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}
	return strings.Split(path, "/")
}

func matchPattern(pattern, path string) (map[string]string, bool) {
	patternParts := splitPath(pattern)
	pathParts := splitPath(path)

	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i, part := range patternParts {
		if len(part) > 0 && part[0] == ':' {
			key := part[1:]
			params[key] = pathParts[i]
		} else if part != pathParts[i] {
			return nil, false
		}
	}

	return params, true
}
