package main

import (
	"encoding/json"
	"fmt"
)

func CommandResponse(id string, data interface{}, err error) [][]byte {
	if err != nil {
		return [][]byte{ErrorResponse(id, err)}
	}
	return [][]byte{ObjectResponse(id, data)}
}

// ErrorResponse serializes an error into response bytes
func ErrorResponse(id string, err error) []byte {
	b, err := json.Marshal(map[string]string{
		"id":    id,
		"error": err.Error(),
	})
	if err != nil {
		panic(fmt.Errorf("error response with error: %s", err.Error()))
	}
	return b
}

// ObjectResponse serializes the given object into response bytes
func ObjectResponse(id string, object interface{}) []byte {
	b, err := json.Marshal(map[string]interface{}{
		"id":   id,
		"data": object,
	})
	if err != nil {
		panic(fmt.Errorf("json response with error: %s", err.Error()))
	}
	return b
}
