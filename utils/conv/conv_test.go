package conv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSP(t *testing.T) {
	assert.Equal(t, "", *SP(""))
	assert.Equal(t, " ", *SP(" "))
	assert.Equal(t, "a", *SP("a"))
	assert.Equal(t, "foo", *SP("foo"))
	assert.Equal(t, "foo bar foo bar", *SP("foo bar foo bar"))
}

func TestS(t *testing.T) {
	assert.Equal(t, "", S(nil))
	assert.Equal(t, "", S(SP("")))
	assert.Equal(t, " ", S(SP(" ")))
	assert.Equal(t, "   ", S(SP("   ")))
	assert.Equal(t, "a", S(SP("a")))
	assert.Equal(t, "foo", S(SP("foo")))
	assert.Equal(t, "foo bar", S(SP("foo bar")))
}
