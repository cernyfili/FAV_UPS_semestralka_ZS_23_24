package command_processing_utils

import (
	"fmt"
	"gameserver/internal/models"
	"gameserver/internal/utils/errorHandeling"
)

func ProcessResponseClientSucessByPlayer(player *models.Player, timeStamp string) error {
	err := player.DecreaseResponseSuccessExpected(timeStamp)
	if err != nil {
		err = fmt.Errorf("Error decreasing response expected: %w", err)
		errorHandeling.PrintError(err)
		return err
	}

	return nil
}
