package main

import (
	"encoding/json"
	"fmt"
)

// ErrorResponse serializes an error into response bytes
func ErrorResponse(err error) []byte {
	b, err := json.Marshal(map[string]string{
		"error": err.Error(),
	})
	if err != nil {
		panic(fmt.Errorf("error response with error: %s", err.Error()))
	}
	return b
}

// ObjectResponse serializes the given object into response bytes
func ObjectResponse(object interface{}) []byte {
	b, err := json.Marshal(map[string]interface{}{
		"data": object,
	})
	if err != nil {
		panic(fmt.Errorf("json response with error: %s", err.Error()))
	}
	return b
}
