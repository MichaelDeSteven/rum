package rum

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateEngine(t *testing.T) {
	router := Deafult()
	assert.Equal(t, ":9999", router.addr)
}

func TestMethodGet(t *testing.T) {
	url := "./testdata/template/hello.tmpl"
	e := Deafult()
	e.Start()
	e.GET("/test", HandlersChain{func(c *Context) {}})

	res, err := http.Get(fmt.Sprintf("%s/test", url))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%+v\n", res)
	resp, _ := ioutil.ReadAll(res.Body)
	assert.Equal(t, "<h1>Hello world</h1>", string(resp))
}
