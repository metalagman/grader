package hw1

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFoo(t *testing.T) {
	v := Foo()
	assert.Equal(t, "hello world!", v)
}
