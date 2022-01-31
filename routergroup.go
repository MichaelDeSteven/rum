package rum

import "net/http"

// IRouter defines all router handle interface includes single and group router.
type IRouter interface {
	IRoutes
	Group(string, ...HandlerFunc) *RouterGroup
}

// IRoutes defines all router handle interface.
type IRoutes interface {
	Use(...HandlerFunc) IRoutes

	Handle(string, string, ...HandlerFunc) IRoutes
	GET(string, ...HandlerFunc) IRoutes
	POST(string, ...HandlerFunc) IRoutes
	DELETE(string, ...HandlerFunc) IRoutes
	PUT(string, ...HandlerFunc) IRoutes
}

type RouterGroup struct {
	engine *Engine

	BasePath string

	Handlers HandlersChain

	root bool
}

func (group *RouterGroup) combine(handlers HandlersChain) HandlersChain {
	finalSize := len(group.Handlers) + len(handlers)
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

// add middleware to RouterGroup
func (group *RouterGroup) Use(handlers ...HandlerFunc) IRoutes {
	group.Handlers = append(group.Handlers, handlers...)
	return group.returnObj()
}

// create a new RouterGroup
func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	engine := group.engine

	return &RouterGroup{
		BasePath: joinPath(group.BasePath, relativePath),
		engine:   engine,
		Handlers: group.combine(handlers),
	}
}

func (group *RouterGroup) GET(path string, handlers ...HandlerFunc) IRoutes {
	return group.handle(http.MethodGet, path, handlers)
}

func (group *RouterGroup) POST(path string, handlers ...HandlerFunc) IRoutes {
	return group.handle(http.MethodPost, path, handlers)
}

func (group *RouterGroup) DELETE(path string, handlers ...HandlerFunc) IRoutes {
	return group.handle(http.MethodDelete, path, handlers)
}

func (group *RouterGroup) PUT(path string, handlers ...HandlerFunc) IRoutes {
	return group.handle(http.MethodPut, path, handlers)
}

func (group *RouterGroup) HEAD(path string, handlers ...HandlerFunc) IRoutes {
	return group.handle(http.MethodHead, path, handlers)
}

func (group *RouterGroup) Handle(httpMethod, path string, handlers ...HandlerFunc) IRoutes {
	return group.handle(httpMethod, path, handlers)
}

func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) IRoutes {
	absolutePath := joinPath(group.BasePath, relativePath)
	handlers = group.combine(handlers)
	group.engine.addRoute(httpMethod, absolutePath, handlers)
	return group.returnObj()
}

func (group *RouterGroup) returnObj() IRoutes {
	if group.root {
		return group.engine
	}
	return group
}
