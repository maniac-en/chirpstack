package utils

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestRemoveProfanity(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
		expectedClean  bool
	}{
		{
			name:           "no profanity",
			input:          "This is a clean message",
			expectedOutput: "This is a clean message",
			expectedClean:  false,
		},
		{
			name:           "single profane word",
			input:          "This is kerfuffle",
			expectedOutput: "This is ****",
			expectedClean:  true,
		},
		{
			name:           "multiple profane words",
			input:          "kerfuffle and sharbert are bad",
			expectedOutput: "**** and **** are bad",
			expectedClean:  true,
		},
		{
			name:           "profane word with different case",
			input:          "KERFUFFLE is loud",
			expectedOutput: "**** is loud",
			expectedClean:  true,
		},
		{
			name:           "mixed case profanity",
			input:          "KerFuFfLe is mixed",
			expectedOutput: "**** is mixed",
			expectedClean:  true,
		},
		{
			name:           "all profane words",
			input:          "kerfuffle sharbert fornax",
			expectedOutput: "**** **** ****",
			expectedClean:  true,
		},
		{
			name:           "empty string",
			input:          "",
			expectedOutput: "",
			expectedClean:  false,
		},
		{
			name:           "profanity with extra spaces",
			input:          "this  is  kerfuffle  word",
			expectedOutput: "this is **** word",
			expectedClean:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, cleaned := RemoveProfanity(tt.input)
			if output != tt.expectedOutput {
				t.Errorf("RemoveProfanity() output = %q, want %q", output, tt.expectedOutput)
			}
			if cleaned != tt.expectedClean {
				t.Errorf("RemoveProfanity() cleaned = %v, want %v", cleaned, tt.expectedClean)
			}
		})
	}
}

func TestRespondWithJSON(t *testing.T) {
	tests := []struct {
		name           string
		code           int
		payload        any
		expectedCode   int
		expectedBody   string
		expectedHeader string
		wantErr        bool
	}{
		{
			name:           "valid JSON response",
			code:           200,
			payload:        map[string]string{"message": "success"},
			expectedCode:   200,
			expectedBody:   `{"message":"success"}`,
			expectedHeader: "application/json",
			wantErr:        false,
		},
		{
			name:           "status created",
			code:           201,
			payload:        map[string]int{"id": 123},
			expectedCode:   201,
			expectedBody:   `{"id":123}`,
			expectedHeader: "application/json",
			wantErr:        false,
		},
		{
			name:           "nil payload",
			code:           204,
			payload:        nil,
			expectedCode:   204,
			expectedBody:   "null",
			expectedHeader: "application/json",
			wantErr:        false,
		},
		{
			name:           "empty struct",
			code:           200,
			payload:        struct{}{},
			expectedCode:   200,
			expectedBody:   "{}",
			expectedHeader: "application/json",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := RespondWithJSON(w, tt.code, tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("RespondWithJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if w.Code != tt.expectedCode {
				t.Errorf("RespondWithJSON() code = %v, want %v", w.Code, tt.expectedCode)
			}

			if w.Header().Get("Content-Type") != tt.expectedHeader {
				t.Errorf("RespondWithJSON() header = %v, want %v", w.Header().Get("Content-Type"), tt.expectedHeader)
			}

			if w.Body.String() != tt.expectedBody {
				t.Errorf("RespondWithJSON() body = %v, want %v", w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestRespondWithError(t *testing.T) {
	tests := []struct {
		name         string
		code         int
		message      string
		expectedCode int
		expectedBody string
	}{
		{
			name:         "bad request error",
			code:         400,
			message:      "Bad Request",
			expectedCode: 400,
			expectedBody: `{"error":"Bad Request"}`,
		},
		{
			name:         "unauthorized error",
			code:         401,
			message:      "Unauthorized",
			expectedCode: 401,
			expectedBody: `{"error":"Unauthorized"}`,
		},
		{
			name:         "not found error",
			code:         404,
			message:      "Not Found",
			expectedCode: 404,
			expectedBody: `{"error":"Not Found"}`,
		},
		{
			name:         "internal server error",
			code:         500,
			message:      "Internal Server Error",
			expectedCode: 500,
			expectedBody: `{"error":"Internal Server Error"}`,
		},
		{
			name:         "empty message",
			code:         400,
			message:      "",
			expectedCode: 400,
			expectedBody: `{"error":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := RespondWithError(w, tt.code, tt.message)
			if err != nil {
				t.Errorf("RespondWithError() unexpected error = %v", err)
			}

			if w.Code != tt.expectedCode {
				t.Errorf("RespondWithError() code = %v, want %v", w.Code, tt.expectedCode)
			}

			if w.Header().Get("Content-Type") != "application/json" {
				t.Errorf("RespondWithError() content-type = %v, want application/json", w.Header().Get("Content-Type"))
			}

			if w.Body.String() != tt.expectedBody {
				t.Errorf("RespondWithError() body = %v, want %v", w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestRespondWithJSONMarshalError(t *testing.T) {
	w := httptest.NewRecorder()

	// Create a value that cannot be marshaled to JSON
	invalidPayload := make(chan int)

	err := RespondWithJSON(w, 200, invalidPayload)
	if err == nil {
		t.Error("RespondWithJSON() expected error for invalid payload but got none")
	}
}

func TestRespondWithJSONAndError_Integration(t *testing.T) {
	// Test that both functions work together correctly
	w := httptest.NewRecorder()

	// First, test successful JSON response
	err := RespondWithJSON(w, 200, map[string]string{"status": "ok"})
	if err != nil {
		t.Fatalf("RespondWithJSON() unexpected error = %v", err)
	}

	// Reset recorder for error response
	w = httptest.NewRecorder()

	// Test error response
	err = RespondWithError(w, 400, "validation failed")
	if err != nil {
		t.Fatalf("RespondWithError() unexpected error = %v", err)
	}

	// Verify error response structure
	var errorResponse map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &errorResponse); err != nil {
		t.Fatalf("Failed to unmarshal error response: %v", err)
	}

	if errorResponse["error"] != "validation failed" {
		t.Errorf("Error response message = %v, want 'validation failed'", errorResponse["error"])
	}
}
