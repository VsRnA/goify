package goify

type RouterGroup struct {
	router *Router
	prefix string
	middleware []MiddlewareFunc
}

func (rt *Router) Group (prefix string) *RouterGroup {
	return &RouterGroup{
		router:	rt,
		prefix: prefix,
		middleware: make([]MiddlewareFunc, 0),
	}
}

func (rg *RouterGroup) Group (prefix string) *RouterGroup {
	return &RouterGroup{
		router:	rg.router,
		prefix: rg.prefix + prefix,
		middleware: append([]MiddlewareFunc{}, rg.middleware...),
	}
}

func (rg *RouterGroup) Use(middleware ...MiddlewareFunc) {
	rg.middleware = append(rg.middleware, middleware...)
}

func (rg *RouterGroup) GET(path string, handler HandlerFunc) {
	rg.addRoute("GET", path, handler)
}

func (rg *RouterGroup) POST(path string, handler HandlerFunc) {
	rg.addRoute("POST", path, handler)
}

func (rg *RouterGroup) PUT(path string, handler HandlerFunc) {
	rg.addRoute("PUT", path, handler)
}

func (rg *RouterGroup) DELETE(path string, handler HandlerFunc) {
	rg.addRoute("DELETE", path, handler)
}

func (rg *RouterGroup) PATCH(path string, handler HandlerFunc) {
	rg.addRoute("PATCH", path, handler)
}

func (rg *RouterGroup) addRoute(method, path string, handler HandlerFunc) {
	fullPath := rg.prefix + path

	wrappedHandler := func(c *Context) {
		rg.executeGroupMiddleware(c, handler)
	}
	
	rg.router.addRoute(method, fullPath, wrappedHandler)
}

func (rg *RouterGroup) executeGroupMiddleware(ctx *Context, handler HandlerFunc) {
	index := 0
	
	var next func()
	next = func() {
		if index < len(rg.middleware) {
			middleware := rg.middleware[index]
			index++
			middleware(ctx, next)
		} else {
			handler(ctx)
		}
	}
	
	next()
}
