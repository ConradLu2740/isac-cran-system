package errors

import (
	"net/http"
	"testing"
)

func TestCode_Message(t *testing.T) {
	tests := []struct {
		code Code
		want string
	}{
		{CodeSuccess, "success"},
		{CodeBadRequest, "bad request"},
		{CodeNotFound, "not found"},
		{CodeInternalError, "internal server error"},
	}

	for _, tt := range tests {
		got := tt.code.Message()
		if got != tt.want {
			t.Errorf("Code(%d).Message() = %q, want %q", tt.code, got, tt.want)
		}
	}
}

func TestCode_Int(t *testing.T) {
	code := CodeBadRequest
	if code.Int() != 400 {
		t.Errorf("CodeBadRequest.Int() = %d, want 400", code.Int())
	}
}

func TestAppError_Error(t *testing.T) {
	err := New(CodeBadRequest, "invalid parameter")

	if err.Code != CodeBadRequest {
		t.Errorf("Expected CodeBadRequest, got %d", err.Code)
	}

	if err.Message != "invalid parameter" {
		t.Errorf("Expected message 'invalid parameter', got %q", err.Message)
	}
}

func TestAppError_ErrorWithDetail(t *testing.T) {
	err := NewWithDetail(CodeBadRequest, "invalid parameter", "field 'name' is required")

	if err.Detail != "field 'name' is required" {
		t.Errorf("Detail = %q, want %q", err.Detail, "field 'name' is required")
	}
}

func TestAppError_Wrap(t *testing.T) {
	innerErr := New(CodeDBConnectError, "connection refused")
	err := Wrap(CodeInternalError, "failed to connect database", innerErr)

	if err.Err != innerErr {
		t.Error("Expected wrapped error")
	}
}

func TestAppError_HTTPStatus(t *testing.T) {
	tests := []struct {
		code     Code
		expected int
	}{
		{CodeUnauthorized, http.StatusUnauthorized},
		{CodeForbidden, http.StatusForbidden},
		{CodeNotFound, http.StatusNotFound},
		{CodeInternalError, http.StatusInternalServerError},
		{CodeIRSDeviceError, http.StatusBadGateway},
		{CodeDBConnectError, http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		err := &AppError{Code: tt.code}
		got := err.HTTPStatus()
		if got != tt.expected {
			t.Errorf("Code %d HTTPStatus() = %d, want %d", tt.code, got, tt.expected)
		}
	}
}

func TestIsCode(t *testing.T) {
	err := New(CodeBadRequest, "test")

	if !IsCode(err, CodeBadRequest) {
		t.Error("Expected IsCode to return true")
	}

	if IsCode(err, CodeNotFound) {
		t.Error("Expected IsCode to return false for different code")
	}
}

func TestGetCode(t *testing.T) {
	err := New(CodeBadRequest, "test")

	if GetCode(err) != CodeBadRequest {
		t.Errorf("GetCode() = %d, want %d", GetCode(err), CodeBadRequest)
	}

	if GetCode(nil) != CodeSuccess {
		t.Errorf("GetCode(nil) = %d, want %d", GetCode(nil), CodeSuccess)
	}
}

func TestBadRequest(t *testing.T) {
	err := BadRequest("invalid input")

	if err.Code != CodeBadRequest {
		t.Errorf("Expected CodeBadRequest, got %d", err.Code)
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("resource not found")

	if err.Code != CodeNotFound {
		t.Errorf("Expected CodeNotFound, got %d", err.Code)
	}
}

func TestInternalError(t *testing.T) {
	innerErr := New(CodeDBConnectError, "connection failed")
	err := InternalError("database error", innerErr)

	if err.Code != CodeInternalError {
		t.Errorf("Expected CodeInternalError, got %d", err.Code)
	}

	if err.Err != innerErr {
		t.Error("Expected inner error to be wrapped")
	}
}
