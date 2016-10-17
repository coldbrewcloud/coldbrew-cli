package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBlank(t *testing.T) {
	assert.True(t, IsBlank(""))
	assert.True(t, IsBlank(" "))
	assert.True(t, IsBlank("    "))
	assert.True(t, IsBlank("\n"))
	assert.True(t, IsBlank("\t"))
	assert.True(t, IsBlank("\t "))
	assert.True(t, IsBlank("\n "))
	assert.True(t, IsBlank("\n \t"))

	assert.False(t, IsBlank("a"))
	assert.False(t, IsBlank("a "))
	assert.False(t, IsBlank(" a"))
	assert.False(t, IsBlank("      a"))
	assert.False(t, IsBlank("\ta"))
	assert.False(t, IsBlank("\na"))
}
