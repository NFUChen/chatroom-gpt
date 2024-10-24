package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

//func SafeRemoveMapKey[K int | string](lock *sync.Mutex, _map *map[K]any, key K) {
//	lock.Lock()
//	defer lock.Unlock()
//	delete(*_map, key)
//}
//
//func SafeAddMapKey[K int | string, V any](lock *sync.Mutex, _map *map[K]any, key K, value *V) {
//	lock.Lock()
//	defer lock.Unlock()
//	(*_map)[key] = value
//}

type ServerResponse[T any] struct {
	Message T   `json:"message"`
	Status  int `json:"status"`
}

const (
	StatusSuccessMin = 200
	StatusSuccessMax = 299
)

func HandleResponse[T any](httpResponse *http.Response) (*ServerResponse[T], error) {
	statusCode := httpResponse.StatusCode
	responseRead, _ := io.ReadAll(httpResponse.Body)
	log.Println("Server response body:", string(responseRead))

	isSuccessful := statusCode >= StatusSuccessMin && statusCode <= StatusSuccessMax
	// Check if the status code indicates success (200-299).
	if !isSuccessful {
		return nil, errors.New(fmt.Sprintf("http status code %d, error: %v", statusCode, string(responseRead)))
	}
	var response ServerResponse[T]
	err := json.Unmarshal(responseRead, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
