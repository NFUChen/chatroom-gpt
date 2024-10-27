package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

const BaseURL = "http://0.0.0.0:8085/api/internal"

const (
	PostSaveMessageApi = BaseURL + "/send_chat_message"
)

const (
	PostAddNewRoom = BaseURL + "/chat_room"
)

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
