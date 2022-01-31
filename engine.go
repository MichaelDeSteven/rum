package rum

import (
	"net/http"
)

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc func(*Context)

// HandlersChain defines a HandlerFunc slice.
type HandlersChain []HandlerFunc

type methodTree struct {
	method string
	root   *node
}

type trees []methodTree

func (trees trees) get(method string) *node {
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

type Engine struct {
	addr string

	trees
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r)
	e.handle(c)
}

func (e *Engine) addRoute(method, path string, handlers HandlersChain) {
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(method != "", "HTTP method can not be empty")
	assert1(len(handlers) > 0, "there must be at least one handler")

	root := e.trees.get(method)
	if root == nil {
		root = new(node)
		e.trees = append(e.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)
}

func (engine *Engine) GET(path string, handlers HandlersChain) {
	engine.addRoute(http.MethodGet, path, handlers)
}

func (engine *Engine) POST(path string, handlers HandlersChain) {
	engine.addRoute(http.MethodPost, path, handlers)
}

func (engine *Engine) DELETE(path string, handlers HandlersChain) {
	engine.addRoute(http.MethodDelete, path, handlers)
}

func (engine *Engine) PUT(path string, handlers HandlersChain) {
	engine.addRoute(http.MethodPut, path, handlers)
}

func (engine *Engine) HEAD(path string, handlers HandlersChain) {
	engine.addRoute(http.MethodHead, path, handlers)
}

func New(addr string) *Engine {
	return &Engine{
		addr:  addr,
		trees: make(trees, 0),
	}
}

func Deafult() *Engine {
	return &Engine{
		addr:  ":9678",
		trees: make(trees, 0),
	}
}

func (e *Engine) Start() {
	http.ListenAndServe(e.addr, e)
}

func (e *Engine) handle(c *Context) {
	handlers, params := e.trees.get(c.Method).getValue(c.Path, &Params{})
	if handlers == nil {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	} else {
		if params != nil {
			c.Params = *params
		}
		handlers[0](c)
	}
}
