package console

import (
	"fmt"

	"github.com/d5/cc"
)

type colorFn func(s string, a ...interface{}) string

func regularFn(s string, a ...interface{}) string {
	return fmt.Sprintf(s, a...)
}

func concat(fns ...colorFn) colorFn {
	return func(s string, a ...interface{}) string {
		out := fmt.Sprintf(s, a...)
		for _, fn := range fns {
			out = fn(out)
		}
		return out
	}
}

var (
	ColorFnAskQuestionNote         = cc.BlackH
	ColorFnAskQuestionMain         = regularFn
	ColorFnAskQuestionDefaultValue = cc.YellowH

	ColorFnAskConfirmNote          = cc.BlackH
	ColorFnAskConfirmMain          = regularFn
	ColorFnAskConfirmDefaultAnswer = regularFn
	ColorFnAskConfirmAnswer        = regularFn

	ColorFnInfoMessage      = regularFn
	ColorFnDetailMessage    = cc.BlackH
	ColorFnSideNote         = cc.BlackH
	ColorFnSideNoteNegative = cc.Red

	ColorFnResource         = cc.Green
	ColorFnResourceNegative = cc.Red

	ColorFnErrorHeader  = cc.Red
	ColorFnErrorMessage = regularFn

	//ColorFnShellCommand = concat(cc.Bold, cc.YellowH)
	ColorFnShellCommand = cc.Cyan
	ColorFnShellOutput  = cc.BlackH
	ColorFnShellError   = cc.Red

	ColorFnMarkAdd      = cc.Green
	ColorFnMarkRemove   = cc.Red
	ColorFnMarkUpdate   = cc.BlueH
	ColorFnMarkProcess  = cc.BlueH
	ColorFnMarkQuestion = cc.BlackH
	ColorFnMarkShell    = regularFn
)

var (
	MarkAdd        = ColorFnMarkAdd("[+]")
	MarkRemove     = ColorFnMarkRemove("[-]")
	MarkUpdate     = ColorFnMarkUpdate("[*]")
	MarkProcessing = ColorFnMarkProcess("[*]")
	MarkQuestion   = ColorFnMarkQuestion(">")
	MarkShell      = ColorFnMarkShell(">")
)
