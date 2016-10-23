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

	ColorFnMarkAdd        = cc.Green
	ColorFnMarkRemove     = cc.Red
	ColorFnMarkUpdate     = cc.BlueH
	ColorFnMarkProcessing = cc.BlueH
	ColorFnMarkQuestion   = cc.BlackH
	ColorFnMarkShell      = regularFn
)

var (
	MarkAdd        = "[+]"
	MarkRemove     = "[-]"
	MarkUpdate     = "[*]"
	MarkProcessing = "[*]"
	MarkQuestion   = ">"
	MarkShell      = ">"
)

func Blank() {
	printfFn("\n")
}

func Info(message string) {
	printfFn("%s\n", ColorFnInfoMessage(message))
}

func DetailWithResource(message, resourceName string) {
	//Println("  " +
	//	ColorFnDetailMessage(message+" [") +
	//	ColorFnResource(resourceName) +
	//	ColorFnDetailMessage("]"))
	printfFn("  %s %s\n", ColorFnDetailMessage(message+":"), ColorFnResource(resourceName))
}

func DetailWithResourceNote(message, resourceName, note string, negative bool) {
	sideNote := ""
	if note != "" {
		if negative {
			sideNote = ColorFnSideNoteNegative(note)
		} else {
			sideNote = ColorFnSideNote(note)
		}
	}

	//Printf("  %s%s%s %s\n",
	//	ColorFnDetailMessage(message+" ["),
	//	ColorFnResource(resourceName),
	//	ColorFnDetailMessage("]"),
	//	sideNote)
	printfFn("  %s %s %s\n",
		ColorFnDetailMessage(message+":"),
		ColorFnResource(resourceName),
		sideNote)
}

func AddingResource(message, resourceName string, mayTakeLong bool) {
	sideNote := ""
	if mayTakeLong {
		sideNote = ColorFnSideNote("(this may take long)")
	}

	printfFn("%s %s%s%s... %s\n",
		ColorFnMarkAdd(MarkAdd),
		ColorFnInfoMessage(message+" ["),
		ColorFnResource(resourceName),
		ColorFnInfoMessage("]"),
		sideNote)

}

func RemovingResource(message, resourceName string, mayTakeLong bool) {
	sideNote := ""
	if mayTakeLong {
		sideNote = ColorFnSideNote("(this may take long)")
	}

	printfFn("%s %s%s%s... %s\n",
		ColorFnMarkRemove(MarkRemove),
		ColorFnInfoMessage(message+" ["),
		ColorFnResourceNegative(resourceName),
		ColorFnInfoMessage("]"),
		sideNote)
}

func UpdatingResource(message, resourceName string, mayTakeLong bool) {
	sideNote := ""
	if mayTakeLong {
		sideNote = ColorFnSideNote("(this may take long)")
	}

	printfFn("%s %s%s%s... %s\n",
		ColorFnMarkUpdate(MarkUpdate),
		ColorFnInfoMessage(message+" ["),
		ColorFnResource(resourceName),
		ColorFnInfoMessage("]"),
		sideNote)
}

func ProcessingOnResource(message, resourceName string, mayTakeLong bool) {
	sideNote := ""
	if mayTakeLong {
		sideNote = ColorFnSideNote("(this may take long)")
	}

	printfFn("%s %s%s%s... %s\n",
		ColorFnMarkProcessing(MarkProcessing),
		ColorFnInfoMessage(message+" ["),
		ColorFnResource(resourceName),
		ColorFnInfoMessage("]"),
		sideNote)
}

func ShellCommand(message string) {
	printfFn("%s %s\n",
		ColorFnMarkShell(MarkShell),
		ColorFnShellCommand(message))
}

func ShellOutput(message string) {
	printfFn("%s\n", ColorFnShellOutput(message))
}

func ShellError(message string) {
	printfFn("%s\n", ColorFnShellError(message))
}
