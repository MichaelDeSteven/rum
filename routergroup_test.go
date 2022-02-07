package rum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouterGroupBasic(t *testing.T) {
	router := New("9678")
	group := router.Group("/hola", func(c *Context) {})
	group.Use(func(c *Context) {})

	assert.Len(t, group.Handlers, 2)
	assert.Equal(t, "/hola", group.BasePath)
	assert.Equal(t, router, group.engine)

	group2 := group.Group("manu")
	group2.Use(func(c *Context) {}, func(c *Context) {})

	assert.Len(t, group2.Handlers, 4)
	assert.Equal(t, "/hola/manu", group2.BasePath)
	assert.Equal(t, router, group2.engine)
}
