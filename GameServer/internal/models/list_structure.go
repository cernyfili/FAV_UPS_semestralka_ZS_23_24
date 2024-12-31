package models

import (
	"fmt"
	"sync"
)

type List struct {
	data  map[interface{}]interface{}
	Count int
	mutex sync.Mutex
}

// region FUNCTIONS
func CreateList() *List {
	return &List{
		data:  make(map[interface{}]interface{}),
		Count: 0,
	}
}

// region ADD
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

//endregion

// region GETTERS
func (gm *List) GetItem(key interface{}) (interface{}, error) {
	// Check if the key exists
	if _, ok := gm.data[key]; !ok {
		return nil, fmt.Errorf("key does not exist")
	}

	// Return the item
	return gm.data[key], nil
}

// GetValuesArray returns an array of all the values in the list
func (gm *List) GetValuesArray() []interface{} {
	var values []interface{}
	for _, v := range gm.data {
		values = append(values, v)
	}
	return values
}

//endregion

// RemoveItem removes an item from the list
func (gm *List) RemoveItem(key interface{}) error {
	// Check if the key exists
	if _, ok := gm.data[key]; !ok {
		return fmt.Errorf("key does not exist")
	}

	// Remove the item from the map
	delete(gm.data, key)
	gm.Count--

	return nil
}

func (gm *List) HasValue(value interface{}) bool {
	for _, v := range gm.data {
		if v == value {
			return true
		}
	}
	return false
}

//endregion
