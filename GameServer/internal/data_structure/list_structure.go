package data_structure

import "errors"

type List struct {
	data  map[int]interface{}
	Count int
}

func CreateList() *List {
	return &List{
		data:  make(map[int]interface{}),
		Count: 0,
	}
}

func (gm *List) AddItem(item interface{}) (int, error) {
	// Generate a unique key based on the current Count
	key := gm.Count

	// Check if the key already exists
	if _, ok := gm.data[key]; ok {
		return -1, errors.New("key already exists")
	}

	// Add the item to the map with the key
	gm.data[key] = item
	gm.Count++

	return key, nil
}

func (gm *List) GetItem(key int) (interface{}, error) {
	// Check if the key exists
	if _, ok := gm.data[key]; !ok {
		return nil, errors.New("key does not exist")
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
