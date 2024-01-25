package utils

import (
	"fmt"
	"sync"
)

type List struct {
	data  map[interface{}]interface{}
	Count int
	mutex sync.Mutex
}

func CreateList() *List {
	return &List{
		data:  make(map[interface{}]interface{}),
		Count: 0,
	}
}

func (gm *List) AddItem(item interface{}) (int, error) {
	// Generate a unique key based on the current Count
	key := gm.Count

	// Check if the key already exists
	if _, ok := gm.data[key]; ok {
		return -1, fmt.Errorf("key already exists")
	}

	// Add the item to the map with the key
	gm.data[key] = item
	gm.Count++

	return key, nil
}

func (gm *List) AddItemKey(key interface{}, item interface{}) error {
	// Check if the key already exists
	if _, ok := gm.data[key]; ok {
		return fmt.Errorf("key already exists")
	}

	// Add the item to the map with the key
	gm.data[key] = item
	gm.Count++

	return nil
}

func (gm *List) GetItem(key interface{}) (interface{}, error) {
	// Check if the key exists
	if _, ok := gm.data[key]; !ok {
		return nil, fmt.Errorf("key does not exist")
	}

	// Return the item
	return gm.data[key], nil
}

func (gm *List) HasValue(value interface{}) bool {
	for _, v := range gm.data {
		if v == value {
			return true
		}
	}
	return false
}
