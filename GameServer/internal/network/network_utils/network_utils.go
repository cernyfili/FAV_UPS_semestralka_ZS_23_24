package network_utils

import "gameserver/internal/models"

var GReceivedMessageList = models.CreateMessageList(models.Received)
var GSendMessageList = models.CreateMessageList(models.Send)
