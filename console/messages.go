package console

func Blank() {
	Println()
}

func Info(message string) {
	Println(ColorFnInfoMessage(message))
}

func DetailWithAWSResourceName(message, resourceName string) {
	Println("  " +
		ColorFnDetailMessage(message+" [") +
		ColorFnAWSResourceName(resourceName) +
		ColorFnDetailMessage("]"))
}

func AddingAWSResourceName(message, resourceName string, mayTakeLong bool) {
	sideNote := ""
	if mayTakeLong {
		sideNote = ColorFnSideNote("(this may take long)")
	}

	Printf("%s %s%s%s... %s\n",
		MarkAdd,
		ColorFnInfoMessage(message+" ["),
		ColorFnAWSResourceName(resourceName),
		ColorFnInfoMessage("]"),
		sideNote)

}

func RemovingAWSResourceName(message, resourceName string, mayTakeLong bool) {
	sideNote := ""
	if mayTakeLong {
		sideNote = ColorFnSideNote("(this may take long)")
	}

	Printf("%s %s%s%s... %s\n",
		MarkRemove,
		ColorFnInfoMessage(message+" ["),
		ColorFnAWSResourceNameNegative(resourceName),
		ColorFnInfoMessage("]"),
		sideNote)
}

func UpdatingAWSResourceName(message, resourceName string, mayTakeLong bool) {
	sideNote := ""
	if mayTakeLong {
		sideNote = ColorFnSideNote("(this may take long)")
	}

	Printf("%s %s%s%s... %s\n",
		MarkUpdate,
		ColorFnInfoMessage(message+" ["),
		ColorFnAWSResourceName(resourceName),
		ColorFnInfoMessage("]"),
		sideNote)
}
