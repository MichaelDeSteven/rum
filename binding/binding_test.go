// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FooStruct struct {
	Foo string `msgpack:"foo" json:"foo" form:"foo" xml:"foo" binding:"required,max=32"`
}

type FooStructUseNumber struct {
	Foo interface{} `json:"foo" binding:"required"`
}

type FooStructDisallowUnknownFields struct {
	Foo interface{} `json:"foo" binding:"required"`
}

func TestBindingDefault(t *testing.T) {
	assert.Equal(t, Form, Default("GET", ""))
	assert.Equal(t, Form, Default("GET", MIMEJSON))

	assert.Equal(t, JSON, Default("POST", MIMEJSON))
	assert.Equal(t, JSON, Default("PUT", MIMEJSON))

	assert.Equal(t, FormMultipart, Default("POST", MIMEMultipartPOSTForm))
	assert.Equal(t, FormMultipart, Default("PUT", MIMEMultipartPOSTForm))
}

func TestBindingJSONNilBody(t *testing.T) {
	var obj FooStruct
	req, _ := http.NewRequest(http.MethodPost, "/", nil)
	err := JSON.Bind(req, &obj)
	assert.Error(t, err)
}

func TestBindingJSON(t *testing.T) {
	testBodyBinding(t,
		JSON, "json",
		"/", "/",
		`{"foo": "bar"}`, `{"bar": "foo"}`)
}

func TestBindingJSONSlice(t *testing.T) {
	EnableDecoderDisallowUnknownFields = true
	defer func() {
		EnableDecoderDisallowUnknownFields = false
	}()

	testBodyBindingSlice(t, JSON, "json", "/", "/", `[]`, ``)
	testBodyBindingSlice(t, JSON, "json", "/", "/", `[{"foo": "123"}]`, `[{}]`)
	testBodyBindingSlice(t, JSON, "json", "/", "/", `[{"foo": "123"}]`, `[{"foo": ""}]`)
	testBodyBindingSlice(t, JSON, "json", "/", "/", `[{"foo": "123"}]`, `[{"foo": 123}]`)
	testBodyBindingSlice(t, JSON, "json", "/", "/", `[{"foo": "123"}]`, `[{"bar": 123}]`)
	testBodyBindingSlice(t, JSON, "json", "/", "/", `[{"foo": "123"}]`, `[{"foo": "123456789012345678901234567890123"}]`)
}

func TestBindingJSONUseNumber(t *testing.T) {
	testBodyBindingUseNumber(t,
		JSON, "json",
		"/", "/",
		`{"foo": 123}`, `{"bar": "foo"}`)
}

func TestBindingJSONUseNumber2(t *testing.T) {
	testBodyBindingUseNumber2(t,
		JSON, "json",
		"/", "/",
		`{"foo": 123}`, `{"bar": "foo"}`)
}

func TestBindingJSONDisallowUnknownFields(t *testing.T) {
	testBodyBindingDisallowUnknownFields(t, JSON,
		"/", "/",
		`{"foo": "bar"}`, `{"foo": "bar", "what": "this"}`)
}

func TestBindingJSONStringMap(t *testing.T) {
	testBodyBindingStringMap(t, JSON,
		"/", "/",
		`{"foo": "bar", "hello": "world"}`, `{"num": 2}`)
}

func testBodyBinding(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	assert.Equal(t, name, b.Name())

	obj := FooStruct{}
	req := requestWithBody("POST", path, body)
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)

	obj = FooStruct{}
	req = requestWithBody("POST", badPath, badBody)
	err = JSON.Bind(req, &obj)
	assert.Error(t, err)
}

func requestWithBody(method, path, body string) (req *http.Request) {
	req, _ = http.NewRequest(method, path, bytes.NewBufferString(body))
	return
}

func testBodyBindingSlice(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	assert.Equal(t, name, b.Name())

	var obj1 []FooStruct
	req := requestWithBody("POST", path, body)
	err := b.Bind(req, &obj1)
	assert.NoError(t, err)

	var obj2 []FooStruct
	req = requestWithBody("POST", badPath, badBody)
	err = JSON.Bind(req, &obj2)
	assert.Error(t, err)
}

func testBodyBindingStringMap(t *testing.T, b Binding, path, badPath, body, badBody string) {
	obj := make(map[string]string)
	req := requestWithBody("POST", path, body)
	if b.Name() == "form" {
		req.Header.Add("Content-Type", MIMEPOSTForm)
	}
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.NotNil(t, obj)
	assert.Len(t, obj, 2)
	assert.Equal(t, "bar", obj["foo"])
	assert.Equal(t, "world", obj["hello"])

	if badPath != "" && badBody != "" {
		obj = make(map[string]string)
		req = requestWithBody("POST", badPath, badBody)
		err = b.Bind(req, &obj)
		assert.Error(t, err)
	}

	objInt := make(map[string]int)
	req = requestWithBody("POST", path, body)
	err = b.Bind(req, &objInt)
	assert.Error(t, err)
}

func testBodyBindingUseNumber(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	assert.Equal(t, name, b.Name())

	obj := FooStructUseNumber{}
	req := requestWithBody("POST", path, body)
	EnableDecoderUseNumber = true
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	// we hope it is int64(123)
	v, e := obj.Foo.(json.Number).Int64()
	assert.NoError(t, e)
	assert.Equal(t, int64(123), v)

	obj = FooStructUseNumber{}
	req = requestWithBody("POST", badPath, badBody)
	err = JSON.Bind(req, &obj)
	assert.Error(t, err)
}

func testBodyBindingUseNumber2(t *testing.T, b Binding, name, path, badPath, body, badBody string) {
	assert.Equal(t, name, b.Name())

	obj := FooStructUseNumber{}
	req := requestWithBody("POST", path, body)
	EnableDecoderUseNumber = false
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	// it will return float64(123) if not use EnableDecoderUseNumber
	// maybe it is not hoped
	assert.Equal(t, float64(123), obj.Foo)

	obj = FooStructUseNumber{}
	req = requestWithBody("POST", badPath, badBody)
	err = JSON.Bind(req, &obj)
	assert.Error(t, err)
}

func testBodyBindingDisallowUnknownFields(t *testing.T, b Binding, path, badPath, body, badBody string) {
	EnableDecoderDisallowUnknownFields = true
	defer func() {
		EnableDecoderDisallowUnknownFields = false
	}()

	obj := FooStructDisallowUnknownFields{}
	req := requestWithBody("POST", path, body)
	err := b.Bind(req, &obj)
	assert.NoError(t, err)
	assert.Equal(t, "bar", obj.Foo)

	obj = FooStructDisallowUnknownFields{}
	req = requestWithBody("POST", badPath, badBody)
	err = JSON.Bind(req, &obj)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "what")
}
