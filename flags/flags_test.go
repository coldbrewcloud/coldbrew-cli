package flags

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/alecthomas/kingpin.v2"
)

type nullWriter struct{}

func (w *nullWriter) Write(p []byte) (int, error) {
	return 0, nil
}

func testSptr(s string) *string {
	p := s
	return &p
}

func testU16ptr(u uint16) *uint16 {
	p := u
	return &p
}

func testU64ptr(u uint64) *uint64 {
	p := u
	return &p
}

func testBptr(b bool) *bool {
	p := b
	return &p
}

func testBytePtr(b byte) *byte {
	p := b
	return &p
}

func testArgs(command *string, args ...string) []string {
	newArgs := []string{}
	if command != nil {
		newArgs = append(newArgs, *command)
	}
	return append(newArgs, args...)
}

func testUint16Flag(t *testing.T, app *kingpin.Application, target **uint16, command *string, flag string, short *byte, defaultValue *uint16, envVar *string) {
	var err error

	if envVar != nil {
		os.Setenv(*envVar, "")
	}

	if defaultValue != nil {
		_, err = app.Parse(testArgs(command)) // default
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, *defaultValue, **target)
	}

	_, err = app.Parse(testArgs(command, "--"+flag+"=10")) // set by param (long)
	assert.Nil(t, err)
	assert.NotNil(t, *target)
	assert.Equal(t, uint16(10), **target)

	_, err = app.Parse(testArgs(command, "--"+flag, "20")) // set by param (long)
	assert.Nil(t, err)
	assert.NotNil(t, *target)
	assert.Equal(t, uint16(20), **target)

	if short != nil {
		_, err = app.Parse(testArgs(command, fmt.Sprintf("-%c=30", *short))) // set by param (short)
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, uint16(30), **target)

		_, err = app.Parse(testArgs(command, fmt.Sprintf("-%c", *short), "40")) // set by param (short)
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, uint16(40), **target)
	}

	if envVar != nil {
		os.Setenv(*envVar, "50")
		_, err = app.Parse(testArgs(command)) // set by env var
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, uint16(50), **target)

		os.Setenv(*envVar, "60")
		_, err = app.Parse(testArgs(command, "--"+flag+"=70")) // param overrides env var
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, uint16(70), **target)
	}

	_, err = app.Parse(testArgs(command, "--"+flag+"=")) // invalid
	assert.NotNil(t, err)
	_, err = app.Parse(testArgs(command, "--"+flag)) // invalid
	assert.NotNil(t, err)
	_, err = app.Parse(testArgs(command, "--"+flag+" 80")) // invalid
	assert.NotNil(t, err)
}

func testUint64Flag(t *testing.T, app *kingpin.Application, target **uint64, command *string, flag string, short *byte, defaultValue *uint64, envVar *string) {
	var err error

	if envVar != nil {
		os.Setenv(*envVar, "")
	}

	if defaultValue != nil {
		_, err = app.Parse(testArgs(command)) // default
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, *defaultValue, **target)
	}

	_, err = app.Parse(testArgs(command, "--"+flag+"=10")) // set by param (long)
	assert.Nil(t, err)
	assert.NotNil(t, *target)
	assert.Equal(t, uint64(10), **target)

	_, err = app.Parse(testArgs(command, "--"+flag, "20")) // set by param (long)
	assert.Nil(t, err)
	assert.NotNil(t, *target)
	assert.Equal(t, uint64(20), **target)

	if short != nil {
		_, err = app.Parse(testArgs(command, fmt.Sprintf("-%c=30", *short))) // set by param (short)
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, uint64(30), **target)

		_, err = app.Parse(testArgs(command, fmt.Sprintf("-%c", *short), "40")) // set by param (short)
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, uint64(40), **target)
	}

	if envVar != nil {
		os.Setenv(*envVar, "50")
		_, err = app.Parse(testArgs(command)) // set by env var
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, uint64(50), **target)

		os.Setenv(*envVar, "60")
		_, err = app.Parse(testArgs(command, "--"+flag+"=70")) // param overrides env var
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, uint64(70), **target)
	}

	_, err = app.Parse(testArgs(command, "--"+flag+"=")) // invalid
	assert.NotNil(t, err)
	_, err = app.Parse(testArgs(command, "--"+flag)) // invalid
	assert.NotNil(t, err)
	_, err = app.Parse(testArgs(command, "--"+flag+" 80")) // invalid
	assert.NotNil(t, err)
}

func testStringFlag(t *testing.T, app *kingpin.Application, target **string, command *string, flag string, short *byte, defaultValue *string, envVar *string) {
	var err error

	if envVar != nil {
		os.Setenv(*envVar, "")
	}

	if defaultValue != nil {
		_, err = app.Parse(testArgs(command)) // default
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, *defaultValue, **target)
	}

	_, err = app.Parse(testArgs(command, "--"+flag+"=value1")) // set by param (long)
	assert.Nil(t, err)
	assert.NotNil(t, *target)
	assert.Equal(t, "value1", **target)

	_, err = app.Parse(testArgs(command, "--"+flag, "value2")) // set by param (long)
	assert.Nil(t, err)
	assert.NotNil(t, *target)
	assert.Equal(t, "value2", **target)

	if short != nil {
		_, err = app.Parse(testArgs(command, fmt.Sprintf("-%c=value3", *short))) // set by param (short)
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, "value3", **target)

		_, err = app.Parse(testArgs(command, fmt.Sprintf("-%c", *short), "value4")) // set by param (short)
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, "value4", **target)
	}

	if envVar != nil {
		os.Setenv(*envVar, "value5")
		_, err = app.Parse(testArgs(command)) // set by env var
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, "value5", **target)

		os.Setenv(*envVar, "value6")
		_, err = app.Parse(testArgs(command, "--"+flag+"=value7")) // param overrides env var
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, "value7", **target)
	}

	_, err = app.Parse(testArgs(command, "--"+flag+"=")) // empty (valid but should be handled by app)
	assert.Nil(t, err)
	assert.NotNil(t, *target)
	assert.Equal(t, "", **target)

	_, err = app.Parse(testArgs(command, "--"+flag)) // invalid
	assert.NotNil(t, err)
	_, err = app.Parse(testArgs(command, "--"+flag+" ver3")) // invalid
	assert.NotNil(t, err)
}

func testBoolFlag(t *testing.T, app *kingpin.Application, target **bool, command *string, flag string, short *byte, defaultValue *bool) {
	var err error

	if defaultValue != nil {
		_, err = app.Parse(testArgs(command)) // default
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, *defaultValue, **target)
	}

	_, err = app.Parse(testArgs(command, "--"+flag)) // set by param (long)
	assert.Nil(t, err)
	assert.NotNil(t, *target)
	assert.Equal(t, true, **target)

	if short != nil {
		_, err = app.Parse(testArgs(command, fmt.Sprintf("-%c", *short))) // set by param (short)
		assert.Nil(t, err)
		assert.NotNil(t, *target)
		assert.Equal(t, true, **target)
	}

	// disable
	_, err = app.Parse(testArgs(command, "--no-"+flag)) // set by param (long)
	assert.Nil(t, err)
	assert.NotNil(t, *target)
	assert.Equal(t, false, **target)

	_, err = app.Parse(testArgs(command, "--"+flag+"=")) // invalid
	assert.NotNil(t, err)
	_, err = app.Parse(testArgs(command, "--"+flag+" true")) // invalid
	assert.NotNil(t, err)
}
