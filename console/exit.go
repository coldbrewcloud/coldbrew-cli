package console

import (
	"fmt"
	"os"

	"github.com/coldbrewcloud/coldbrew-cli/core"
	"github.com/d5/cc"
)

func ExitWithErrorString(format string, a ...interface{}) error {
	return ExitWithError(fmt.Errorf(format, a))
}

func ExitWithError(err error) error {
	if ei, ok := err.(*core.Error); ok {
		Errorln(cc.Red("Error:"), ei.Error(), cc.BlackH("(more info: "+ei.ExtraInfo()+")"))
	} else {
		Errorln(cc.Red("Error:"), err.Error())
	}
	os.Exit(100)
	return nil
}
