package rum

import (
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
	assert.Equal(t, ":9678", router.addr)
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
