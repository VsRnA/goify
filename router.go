package goify

import (
	"fmt"
	"net/http"
	"strings"
)

type Router struct {
	routes     map[string]map[string]HandlerFunc
	middleware []MiddlewareFunc
	server     *http.Server
}

type HandlerFunc func(*Context)

func New() *Router {
	return &Router{
		routes:     make(map[string]map[string]HandlerFunc),
		middleware: make([]MiddlewareFunc, 0),
	}
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method := req.Method
	path := req.URL.Path
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	if methodRoutes, exists := rt.routes[method]; exists {
		if handler, found := methodRoutes[path]; found {
			ctx := &Context{
				Request:  req,
				Response: w,
				params:   make(map[string]string),
				store:    make(map[string]interface{}),
			}

			rt.executeMiddleware(ctx, handler)
			return
		}
	}

	http.NotFound(w, req)
}

func (rt *Router) addRoute(method, path string, handler HandlerFunc) {
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}
	
	if rt.routes[method] == nil {
		rt.routes[method] = make(map[string]HandlerFunc)
	}
	rt.routes[method][path] = handler
}

func (rt *Router) GET(path string, handler HandlerFunc) {
	rt.addRoute("GET", path, handler)
}

func (rt *Router) POST(path string, handler HandlerFunc) {
	rt.addRoute("POST", path, handler)
}

func (rt *Router) PUT(path string, handler HandlerFunc) {
	rt.addRoute("PUT", path, handler)
}

func (rt *Router) DELETE(path string, handler HandlerFunc) {
	rt.addRoute("DELETE", path, handler)
}

func (rt *Router) PATCH(path string, handler HandlerFunc) {
	rt.addRoute("PATCH", path, handler)
}

func (rt *Router) Listen(addr string) error {
	rt.server = &http.Server{
		Addr:    addr,
		Handler: rt,
	}
	
	fmt.Printf("🚀 Server started on http://localhost%s\n", addr)
	return rt.server.ListenAndServe()
}

func (rt *Router) Shutdown() error {
	if rt.server != nil {
		return rt.server.Close()
	}
	return nil
}