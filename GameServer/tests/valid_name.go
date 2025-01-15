package main

import (
	"gameserver/internal/utils/constants"
)

func IsValidName(name string) bool {
	// if name is none
	if name == "" {
		return false
	}

	// if name is not numbers and letters
	if !constants.IsAlphaNumeric(name) {
		return false
	}

	return len(name) >= constants.CMessageNameMinChars && len(name) <= constants.CMessageNameMaxChars
}

//
//func main() {
//	print(models.IsValidName("Game1"))
//}
