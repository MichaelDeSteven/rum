package rum

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"rum/binding"
	"sync"
)

// BodyKey indicates a default body bytes key.
const BodyKey = "BodyKey"

// Content-Type MIME of the most common data formats.
const (
	MIMEJSON              = binding.MIMEJSON
	MIMEHTML              = binding.MIMEHTML
	MIMEPlain             = binding.MIMEPlain
	MIMEPOSTForm          = binding.MIMEPOSTForm
	MIMEMultipartPOSTForm = binding.MIMEMultipartPOSTForm
)

type Param struct {
	Key   string
	Value string
}

type Params []Param

type Context struct {
	Writer http.ResponseWriter

	Request *http.Request

	Path string

	Params

	Method string

	StatusCode int

	// current executes handler index, see Next() function
	index int8

	// handler list
	HandlersChain

	// This mutex protects Keys map.
	mu sync.RWMutex

	// Keys is a key/value pair exclusively for the context of each request.
	Keys map[string]interface{}

	// Errors is a list of errors attached to all the handlers/middlewares who used this context.
	Errors errorMsgs
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.Writer = w
	if r != nil {
		c.Request = r
		c.Method = r.Method
		c.Path = r.URL.Path
	}
	c.Params = c.Params[:0]
	c.HandlersChain = nil
	c.Keys = nil
	c.index = -1
}

// Next() used in middleware
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.HandlersChain)) {
		c.HandlersChain[c.index](c)
		c.index++
	}
}

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Context) Set(key string, value interface{}) {
	c.mu.Lock()
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}

	c.Keys[key] = value
	c.mu.Unlock()
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exist it returns (nil, false)
func (c *Context) Get(key string) (value interface{}, exists bool) {
	c.mu.RLock()
	value, exists = c.Keys[key]
	c.mu.RUnlock()
	return
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

func (c *Context) String(code int, format string, str ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, str...)))
}

func (c *Context) ContentType() string {
	return filterFlags(c.requestHeader("Content-Type"))
}

func (c *Context) requestHeader(key string) string {
	return c.Request.Header.Get(key)
}

func (c *Context) Bind(obj interface{}) error {
	b := binding.Default(c.Request.Method, c.ContentType())
	return c.ShouldBindWith(obj, b)
}

func (c *Context) BindJSON(obj interface{}) error {
	return c.ShouldBindWith(obj, binding.JSON)
}

func (c *Context) BindHeader(obj interface{}) error {
	return c.ShouldBindWith(obj, binding.Header)
}

func (c *Context) BindQuery(obj interface{}) error {
	return c.ShouldBindWith(obj, binding.Query)
}

func (c *Context) ShouldBindUri(obj interface{}) error {
	m := make(map[string][]string)
	for _, v := range c.Params {
		m[v.Key] = []string{v.Value}
	}
	return binding.Uri.BindUri(m, obj)
}

// ShouldBindBodyWith is similar with ShouldBindWith, but it stores the request
// body into the context, and reuse when it is called again.
func (c *Context) ShouldBindBodyWith(obj interface{}, bb binding.BindingBody) (err error) {
	var body []byte
	if cb, ok := c.Get(BodyKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	if body == nil {
		body, err = ioutil.ReadAll(c.Request.Body)
		if err != nil {
			return err
		}
		c.Set(BodyKey, body)
	}
	return bb.BindBody(body, obj)
}

func (c *Context) ShouldBindWith(obj interface{}, b binding.Binding) error {
	return b.Bind(c.Request, obj)
}
