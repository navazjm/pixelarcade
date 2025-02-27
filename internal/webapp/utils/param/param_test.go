package param

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestReadID(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedID    int64
		expectedError error
	}{
		{
			name:          "valid ID",
			url:           "/some-path/123",
			expectedID:    123,
			expectedError: nil,
		},
		{
			name:          "invalid ID (non-numeric)",
			url:           "/some-path/abc",
			expectedID:    0,
			expectedError: errors.New("invalid id parameter"),
		},
		{
			name:          "invalid ID (negative)",
			url:           "/some-path/-1",
			expectedID:    0,
			expectedError: errors.New("invalid id parameter"),
		},
		{
			name:          "missing ID",
			url:           "/some-path/",
			expectedID:    0,
			expectedError: errors.New("invalid id parameter"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the router
			router := httprouter.New()
			router.GET("/some-path/:id", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
				// Call ReadID directly without modifying context here
				id, err := ReadID(r)
				if err != nil && err.Error() != tt.expectedError.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				if id != tt.expectedID {
					t.Errorf("expected ID %d, got %d", tt.expectedID, id)
				}
			})

			// Create the request with the test URL
			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create Params manually based on the URL
			params := httprouter.Params{
				httprouter.Param{
					Key:   "id",
					Value: tt.url[len("/some-path/"):], // Extracting the ID directly from the URL string
				},
			}

			// Add params to the request's context without needing httprouter.NewContext
			ctx := req.Context()
			ctx = context.WithValue(ctx, httprouter.ParamsKey, params)
			req = req.WithContext(ctx)

			// Record the response
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
		})
	}
}

func TestInjectID(t *testing.T) {
	tests := []struct {
		name       string
		inputID    int64
		expectedID string
	}{
		{
			name:       "valid positive ID",
			inputID:    123,
			expectedID: "123",
		},
		{
			name:       "zero ID",
			inputID:    0,
			expectedID: "0",
		},
		{
			name:       "negative ID",
			inputID:    -1,
			expectedID: "-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a dummy request
			r, err := http.NewRequest("GET", "/some-path", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			// Inject ID into the request context
			r = InjectID(r, tt.inputID)

			// Retrieve parameters from context
			params, ok := r.Context().Value(httprouter.ParamsKey).(httprouter.Params)
			if !ok {
				t.Fatal("failed to retrieve params from context")
			}

			// Ensure the ID matches expected
			if len(params) == 0 || params[0].Key != "id" || params[0].Value != tt.expectedID {
				t.Errorf("expected ID %q, got %q", tt.expectedID, params[0].Value)
			}
		})
	}
}
