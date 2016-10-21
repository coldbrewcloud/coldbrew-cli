package core

import (
	"os"

	"github.com/coldbrewcloud/coldbrew-cli/console"
	"github.com/d5/cc"
)

func ExitWithError(err error) error {
	console.Errorln(cc.Red("Error:"), err.Error())
	os.Exit(100)
	return nil
}

func ExitWithErrorInfo(err error, infoURL string) error {
	console.Errorln(cc.Red("Error:"), err.Error(), cc.BlackH("(more info: "+infoURL+")"))
	os.Exit(101)
	return nil
}
