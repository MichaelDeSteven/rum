// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package rum

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"rum/binding"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CreateTestContext(w http.ResponseWriter) (c *Context, r *Engine) {
	r = Deafult()
	c = r.allocateContext()
	c.reset(w, nil)
	return
}

func TestContextContentType(t *testing.T) {
	c, _ := CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", nil)
	c.Request.Header.Set("Content-Type", "application/json; charset=utf-8")

	assert.Equal(t, "application/json", c.ContentType())
}

func TestContextAutoBindJSON(t *testing.T) {
	c, _ := CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", bytes.NewBufferString("{\"foo\":\"bar\", \"bar\":\"foo\"}"))
	c.Request.Header.Add("Content-Type", MIMEJSON)

	var obj struct {
		Foo string `json:"foo"`
		Bar string `json:"bar"`
	}
	assert.NoError(t, c.Bind(&obj))
	assert.Equal(t, "foo", obj.Bar)
	assert.Equal(t, "bar", obj.Foo)
	assert.Empty(t, c.Errors)
}

func TestContextBindHeader(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := CreateTestContext(w)

	c.Request, _ = http.NewRequest("POST", "/", nil)
	c.Request.Header.Add("rate", "8000")
	c.Request.Header.Add("domain", "music")
	c.Request.Header.Add("limit", "1000")

	var testHeader struct {
		Rate   int    `header:"Rate"`
		Domain string `header:"Domain"`
		Limit  int    `header:"limit"`
	}

	assert.NoError(t, c.BindHeader(&testHeader))
	assert.Equal(t, 8000, testHeader.Rate)
	assert.Equal(t, "music", testHeader.Domain)
	assert.Equal(t, 1000, testHeader.Limit)
	assert.Equal(t, 0, w.Body.Len())
}

func TestContextBindWithQuery(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := CreateTestContext(w)

	c.Request, _ = http.NewRequest("POST", "/?foo=bar&bar=foo", bytes.NewBufferString("foo=unused"))

	var obj struct {
		Foo string `form:"foo"`
		Bar string `form:"bar"`
	}
	assert.NoError(t, c.BindQuery(&obj))
	assert.Equal(t, "foo", obj.Bar)
	assert.Equal(t, "bar", obj.Foo)
	assert.Equal(t, 0, w.Body.Len())
}

func TestContextShouldBindBodyWith(t *testing.T) {
	type typeA struct {
		Foo string `json:"foo" xml:"foo" binding:"required"`
	}
	type typeB struct {
		Bar string `json:"bar" xml:"bar" binding:"required"`
	}
	for _, tt := range []struct {
		name               string
		bindingA, bindingB binding.BindingBody
		bodyA, bodyB       string
	}{
		{
			name:     "JSON & JSON",
			bindingA: binding.JSON,
			bindingB: binding.JSON,
			bodyA:    `{"foo":"FOO"}`,
			bodyB:    `{"bar":"BAR"}`,
		},
	} {
		t.Logf("testing: %s", tt.name)
		// bodyA to typeA and typeB
		{
			w := httptest.NewRecorder()
			c, _ := CreateTestContext(w)
			c.Request, _ = http.NewRequest(
				"POST", "http://example.com", bytes.NewBufferString(tt.bodyA),
			)
			// When it binds to typeA and typeB, it finds the body is
			// not typeB but typeA.
			objA := typeA{}
			assert.NoError(t, c.ShouldBindBodyWith(&objA, tt.bindingA))
			assert.Equal(t, typeA{"FOO"}, objA)
			objB := typeB{}
			assert.Error(t, c.ShouldBindBodyWith(&objB, tt.bindingB))
			assert.NotEqual(t, typeB{"BAR"}, objB)
		}
		// bodyB to typeA and typeB
		{
			// When it binds to typeA and typeB, it finds the body is
			// not typeA but typeB.
			w := httptest.NewRecorder()
			c, _ := CreateTestContext(w)
			c.Request, _ = http.NewRequest(
				"POST", "http://example.com", bytes.NewBufferString(tt.bodyB),
			)
			objA := typeA{}
			assert.Error(t, c.ShouldBindBodyWith(&objA, tt.bindingA))
			assert.NotEqual(t, typeA{"FOO"}, objA)
			objB := typeB{}
			assert.NoError(t, c.ShouldBindBodyWith(&objB, tt.bindingB))
			assert.Equal(t, typeB{"BAR"}, objB)
		}
	}
}

func TestContextBadAutoBind(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := CreateTestContext(w)

	c.Request, _ = http.NewRequest("POST", "http://example.com", bytes.NewBufferString("\"foo\":\"bar\", \"bar\":\"foo\"}"))
	c.Request.Header.Add("Content-Type", MIMEJSON)
	var obj struct {
		Foo string `json:"foo"`
		Bar string `json:"bar"`
	}

	assert.Error(t, c.Bind(&obj))

	assert.Empty(t, obj.Bar)
	assert.Empty(t, obj.Foo)
}

func TestContextFormFile(t *testing.T) {
	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)
	w, err := mw.CreateFormFile("file", "test")
	if assert.NoError(t, err) {
		_, err = w.Write([]byte("test"))
		assert.NoError(t, err)
	}
	mw.Close()
	c, _ := CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", buf)
	c.Request.Header.Set("Content-Type", mw.FormDataContentType())
	f, err := c.FormFile("file")
	if assert.NoError(t, err) {
		assert.Equal(t, "test", f.Filename)
	}

	assert.NoError(t, c.SaveUploadedFile(f, "test"))
}

func TestContextMultipartForm(t *testing.T) {
	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)
	assert.NoError(t, mw.WriteField("foo", "bar"))
	w, err := mw.CreateFormFile("file", "test")
	if assert.NoError(t, err) {
		_, err = w.Write([]byte("test"))
		assert.NoError(t, err)
	}
	mw.Close()
	c, _ := CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", buf)
	c.Request.Header.Set("Content-Type", mw.FormDataContentType())
	f, err := c.MultipartForm()
	if assert.NoError(t, err) {
		assert.NotNil(t, f)
	}

	assert.NoError(t, c.SaveUploadedFile(f.File["file"][0], "test"))
}

func TestSaveUploadedOpenFailed(t *testing.T) {
	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)
	mw.Close()

	c, _ := CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", buf)
	c.Request.Header.Set("Content-Type", mw.FormDataContentType())

	f := &multipart.FileHeader{
		Filename: "file",
	}
	assert.Error(t, c.SaveUploadedFile(f, "test"))
}

func TestSaveUploadedCreateFailed(t *testing.T) {
	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)
	w, err := mw.CreateFormFile("file", "test")
	if assert.NoError(t, err) {
		_, err = w.Write([]byte("test"))
		assert.NoError(t, err)
	}
	mw.Close()
	c, _ := CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", buf)
	c.Request.Header.Set("Content-Type", mw.FormDataContentType())
	f, err := c.FormFile("file")
	if assert.NoError(t, err) {
		assert.Equal(t, "test", f.Filename)
	}

	assert.Error(t, c.SaveUploadedFile(f, "/"))
}
