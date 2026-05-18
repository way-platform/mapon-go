package mapon

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"
	maponv1 "github.com/way-platform/mapon-go/proto/gen/go/wayplatform/connect/mapon/v1"
)

func TestErrorCodeMapping(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		statusCode int
		wantCode   connect.Code
	}{
		{name: "400 -> InvalidArgument", statusCode: http.StatusBadRequest, wantCode: connect.CodeInvalidArgument},
		{name: "401 -> Unauthenticated", statusCode: http.StatusUnauthorized, wantCode: connect.CodeUnauthenticated},
		{name: "403 -> PermissionDenied", statusCode: http.StatusForbidden, wantCode: connect.CodePermissionDenied},
		{name: "404 -> NotFound", statusCode: http.StatusNotFound, wantCode: connect.CodeNotFound},
		{name: "409 -> AlreadyExists", statusCode: http.StatusConflict, wantCode: connect.CodeAlreadyExists},
		{
			name:       "429 -> ResourceExhausted",
			statusCode: http.StatusTooManyRequests,
			wantCode:   connect.CodeResourceExhausted,
		},
		{name: "500 -> Internal", statusCode: http.StatusInternalServerError, wantCode: connect.CodeInternal},
		{name: "501 -> Unimplemented", statusCode: http.StatusNotImplemented, wantCode: connect.CodeUnimplemented},
		{name: "503 -> Unavailable", statusCode: http.StatusServiceUnavailable, wantCode: connect.CodeUnavailable},
		{
			name:       "504 -> DeadlineExceeded",
			statusCode: http.StatusGatewayTimeout,
			wantCode:   connect.CodeDeadlineExceeded,
		},
		{name: "418 -> Unknown", statusCode: http.StatusTeapot, wantCode: connect.CodeUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte("error body"))
			}))
			defer server.Close()

			client, err := NewClient(context.Background(), WithAPIKey("test-key"), WithRetryCount(0))
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}
			client.baseURL = server.URL

			_, err = client.ListUnits(context.Background(), &maponv1.ListUnitsRequest{})
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			gotCode := connect.CodeOf(err)
			if gotCode != tt.wantCode {
				t.Errorf("status %d: got code %v, want %v", tt.statusCode, gotCode, tt.wantCode)
			}
		})
	}
}
