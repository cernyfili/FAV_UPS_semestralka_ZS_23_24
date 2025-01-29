package errorHandeling

const CDebugMode = true

// function for printing error messages
func PrintError(err error) {
	if err != nil && CDebugMode == true {
		panic(err)
	}
}

func AssertError(err error) {
	if err != nil {
		panic(err)
	}
}
