package finowl

import (
	"fmt"
)

// ErrSummaryNotFound is returned when a summary with the specified ID is not found
type ErrSummaryNotFound struct {
	ID int
}

func (e ErrSummaryNotFound) Error() string {
	return fmt.Sprintf("summary with ID %d not found", e.ID)
}

// ErrUnexpectedStatusCode is returned when the API returns an unexpected status code
type ErrUnexpectedStatusCode struct {
	StatusCode int
}

func (e ErrUnexpectedStatusCode) Error() string {
	return fmt.Sprintf("unexpected status code: %d", e.StatusCode)
}

// ErrMissingSections is returned when the content doesn't contain all required sections
type ErrMissingSections struct {
	MissingSections []string
}

func (e ErrMissingSections) Error() string {
	return fmt.Sprintf("missing sections in content: %v", e.MissingSections)
}

// ErrTwitterPostFailed is returned when posting to Twitter fails
type ErrTwitterPostFailed struct {
	Section string
	Cause   error
}

func (e ErrTwitterPostFailed) Error() string {
	return fmt.Sprintf("failed to post %s tweet: %v", e.Section, e.Cause)
}

// ErrAPIRequestFailed is returned when a request to the API fails
type ErrAPIRequestFailed struct {
	Cause error
}

func (e ErrAPIRequestFailed) Error() string {
	return fmt.Sprintf("API request failed: %v", e.Cause)
}
