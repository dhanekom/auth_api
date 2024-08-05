package helpers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccessResponse(t *testing.T) {
	tests := []struct {
		desc string
		data any
		want string
	}{
		{desc: "nil", data: nil, want: `{"status":"success"}`},
		{desc: "string", data: "some string", want: `{"status":"success","data":"some string"}`},
		{desc: "map", data: map[string]any{"message": "some string"}, want: `{"status":"success","data":{"message":"some string"}}`},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			response := SuccessResponse(test.data)
			b, err := json.Marshal(response)
			if err != nil {
				t.Errorf("unexpected error while marshalling struct to json: %s", err.Error())
			}

			assert.Equal(t, test.want, string(b))
		})
	}
}

func TestFailResponse(t *testing.T) {
	tests := []struct {
		desc string
		data any
		want string
	}{
		{desc: "nil", data: nil, want: `{"status":"fail"}`},
		{desc: "string", data: "some string", want: `{"status":"fail","data":"some string"}`},
		{desc: "map", data: map[string]any{"message": "some string"}, want: `{"status":"fail","data":{"message":"some string"}}`},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			response := FailResponse(test.data)
			b, err := json.Marshal(response)
			if err != nil {
				t.Errorf("unexpected error while marshalling struct to json: %s", err.Error())
			}

			assert.Equal(t, test.want, string(b))
		})
	}
}

func TestErrorResponse(t *testing.T) {
	tests := []struct {
		desc    string
		message string
		want    string
	}{
		{desc: "string", message: "some string", want: `{"status":"error","message":"some string"}`},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			response := ErrorResponse(test.message)
			b, err := json.Marshal(response)
			if err != nil {
				t.Errorf("unexpected error while marshalling struct to json: %s", err.Error())
			}

			assert.Equal(t, test.want, string(b))
		})
	}
}
