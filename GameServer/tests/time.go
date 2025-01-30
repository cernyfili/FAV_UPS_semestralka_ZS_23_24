package main

import (
	"gameserver/internal/utils/constants"
	"time"
)

func main() {

	messageTimeStr := "2025-01-30 8:04:05.000000"
	messageTime, _ := time.Parse(constants.CMessageTimeFormat, messageTimeStr)

	currentTime := time.Now()
	formattedCurrentTimeStr := currentTime.Format(constants.CMessageTimeFormat)
	formattedCurrentTime, _ := time.Parse(constants.CMessageTimeFormat, formattedCurrentTimeStr)

	diff := formattedCurrentTime.Sub(messageTime)

	timeOut := 1 * time.Hour

	isTimeout := diff > timeOut
	print(isTimeout)

}
