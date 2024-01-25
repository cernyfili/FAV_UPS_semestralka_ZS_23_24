package utils

import "sync"

type StateMachine struct {
	list *List
}

var (
	instanceSM *GameList
	onceSM     sync.Once
)

func GetInstanceStateMachine() *GameList {
	onceSM.Do(func() {
		instanceGL = &GameList{
			list: CreateList(),
		}
	})
	return instanceSM
}
