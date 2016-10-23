package console

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
		MarkAdd,
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
		MarkRemove,
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
		MarkUpdate,
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
		MarkProcessing,
		ColorFnInfoMessage(message+" ["),
		ColorFnResource(resourceName),
		ColorFnInfoMessage("]"),
		sideNote)
}

func ShellCommand(message string) {
	printfFn("%s %s\n",
		MarkShell,
		ColorFnShellCommand(message))
}

func ShellOutput(message string) {
	printfFn("%s\n", ColorFnShellOutput(message))
}

func ShellError(message string) {
	printfFn("%s\n", ColorFnShellError(message))
}
