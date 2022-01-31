package rum

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type header struct {
	Key   string
	Value string
}

func TestCreateEngine(t *testing.T) {
	router := Deafult()
	assert.Equal(t, ":9999", router.addr)
}

func TestMethodGet(t *testing.T) {
	url := "./testdata/template/hello.tmpl"
	e := Deafult()
	e.Start()
	e.GET("/test", func(c *Context) {})

	res, err := http.Get(fmt.Sprintf("%s/test", url))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", res)
	resp, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}

// PerformRequest for testing router.
func PerformRequest(r http.Handler, method, path string, headers ...header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for _, h := range headers {
		req.Header.Add(h.Key, h.Value)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
