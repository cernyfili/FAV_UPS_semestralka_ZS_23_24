package errorHandeling

import (
	"fmt"
	"gameserver/internal/logger"
)

const CDebugMode = false

// function for printing error messages
func PrintError(err error) {
	logger.Log.Error(err)
	if err != nil && CDebugMode == true {
		panic(err)
	}
}

func AssertError(err error) {
	errNew := fmt.Errorf("AssertError: %w", err)
	logger.Log.Error(errNew)

	if err != nil {
		panic(errNew)
	}
}
