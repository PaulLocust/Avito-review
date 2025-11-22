package entity

import "fmt"

type ErrorCode string

const (
	ErrorTeamExists   ErrorCode = "TEAM_EXISTS"
	ErrorPRExists     ErrorCode = "PR_EXISTS"
	ErrorPRMerged     ErrorCode = "PR_MERGED"
	ErrorNotAssigned  ErrorCode = "NOT_ASSIGNED"
	ErrorNoCandidate  ErrorCode = "NO_CANDIDATE"
	ErrorNotFound     ErrorCode = "NOT_FOUND"
	ErrorInvalidInput ErrorCode = "INVALID_INPUT"
)

type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e AppError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewAppError(code ErrorCode, message string) AppError {
	return AppError{
		Code:    code,
		Message: message,
	}
}