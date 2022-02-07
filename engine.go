package rum

import (
	"net/http"
	"sync"
)

var once sync.Once
var internalEngine *Engine

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

	group *RouterGroup

	pool sync.Pool
}

func (engine *Engine) allocateContext() *Context {
	return &Context{}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := e.pool.Get().(*Context)
	c.reset(w, r)
	e.handle(c)
	e.pool.Put(c)
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

func (e *Engine) Use(middleware ...HandlerFunc) IRoutes {
	return e.group.Use(middleware...)
}

func (e *Engine) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return e.group.Group(relativePath, handlers...)
}

func (e *Engine) GET(path string, handlers ...HandlerFunc) IRoutes {
	return e.group.GET(path, handlers...)
}

func (e *Engine) POST(path string, handlers ...HandlerFunc) IRoutes {
	return e.group.POST(path, handlers...)
}

func (e *Engine) DELETE(path string, handlers ...HandlerFunc) IRoutes {
	return e.group.DELETE(path, handlers...)
}

func (e *Engine) PUT(path string, handlers ...HandlerFunc) IRoutes {
	return e.group.PUT(path, handlers...)
}

func (e *Engine) HEAD(path string, handlers ...HandlerFunc) IRoutes {
	return e.group.HEAD(path, handlers...)
}

func (e *Engine) Handle(method, path string, handlers ...HandlerFunc) IRoutes {
	return e.group.Handle(method, path, handlers...)
}

func New(addr string) *Engine {
	engine := &Engine{
		addr:  addr,
		trees: make(trees, 0),
		group: &RouterGroup{
			BasePath: "/",
			root:     true,
		},
	}
	engine.group.engine = engine
	engine.pool.New = func() interface{} {
		return &Context{
			Params: Params{},
			index:  -1,
		}
	}
	return engine
}

// Deafult returns an Engine instance with the middleware
func Deafult() *Engine {
	once.Do(func() {
		internalEngine = New("9678")
	})
	return internalEngine
}

func (e *Engine) Start() {
	http.ListenAndServe(e.addr, e)
}

func (e *Engine) handle(c *Context) {
	tree := e.trees.get(c.Method)
	handlers, params := tree.getValue(c.Path, &c.Params)
	c.HandlersChain = handlers
	if params != nil {
		c.Params = *params
	}
	if c.HandlersChain == nil {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	} else {
		c.Next()
	}
}
