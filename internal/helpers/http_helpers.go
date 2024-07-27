package helpers

import (
	"encoding/json"
	"net/http"
)

type jsonResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// SuccessResponse creates a standard JSON API success struct based on the JSend specification
func SuccessResponse(data any) *jsonResponse {
	return &jsonResponse{
		Status: "success",
		Data:   data,
	}
}

// SuccessResponse creates a standard JSON API fail struct based on the JSend specification
func FailResponse(data any) *jsonResponse {
	return &jsonResponse{
		Status: "fail",
		Data:   data,
	}
}

// SuccessResponse creates a standard JSON API error struct based on the JSend specification
func ErrorResponse(message string) *jsonResponse {
	return &jsonResponse{
		Status:  "error",
		Message: message,
	}
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(out)
	return err
}

// func (app *Configs) errorJSON(w http.ResponseWriter, err error, status int) error {
// 	payload := jsonResponse{
// 		Status:  "error",
// 		Message: err.Error(),
// 	}

// 	return app.writeJSON(w, status, payload)
// }

func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1048576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	return dec.Decode(&data)
}
