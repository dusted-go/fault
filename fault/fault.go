package fault

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/dusted-go/fault/stack"
)

// ------
// User Error
// ------

// UserError represents an error that was typically caused by the end user.
// A user error is normally a type of error which an application would like to surface back
// to the user. It could be something like a validation error of some user provided input or
// other errors that would normally result in a 4xx status code in a web application context.
//
// User errors should also contain an error code in addition to the error message.
// This helps the parsing of user errors by external programs which act on behalf of the user.
// For example a UserError returned by a HTTP API would return an error code alongside the message
// so that the calling client can parse the error and decide what to do next.
//
// Error codes should ideally be unique and descriptive strings in order to prevent collision in a larger application.
//
// Examples:
//
//	"MISSING_FIRST_NAME": "Please provide your first name"
//	"INVALID_EMAIL_ADDR": "Please provide a valid email address"
type UserError struct {
	// map of error codes and messages
	errors map[string]string

	// codes is used to preserve the order in which
	// errors are being added, since a map[string]string
	// will iterate in random order.
	codes []string
}

// Add appends an additional user error to the collection of errors.
func (e *UserError) Add(code string, msg string) {
	e.codes = append(e.codes, code)
	e.errors[code] = msg
}

// Addf appends an additional user error to the collection of errors.
func (e *UserError) Addf(code string, format string, a ...interface{}) {
	e.Add(code, fmt.Sprintf(format, a...))
}

func (e *UserError) errorMessage(includeCode bool) string {
	if len(e.errors) == 0 {
		return ""
	}
	prefix := "- "
	if len(e.errors) == 1 {
		prefix = ""
	}
	sb := strings.Builder{}
	for _, k := range e.codes {
		v := e.errors[k]
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		if includeCode {
			sb.WriteString(fmt.Sprintf("%s%s (%s)", prefix, v, k))
		} else {
			sb.WriteString(fmt.Sprintf("%s%s", prefix, v))
		}
	}
	return sb.String()
}

// String returns the error message.
func (e *UserError) String() string {
	return e.Error()
}

// Error will return a string of one or all user errors.
//
// If there is only one user error it will be represented as a single string.
//
//	Example:
//	   Email address is required (MISSING_EMAIL_ADDRESS)
//
// If there are more than one user error (e.g. multiple validation errors)
// then a multi line string resembling a list of errors will be returned.
//
//	Example:
//	   - First name is required (MISSING_FIRST_NAME)
//	   - Last name is required (MISSING_LAST_NAME)
//	   - Invalid email address (INVALID_EMAIL_ADDRESS)
//
// Use FriendlyError() to compute the same string without error codes attached.
//
// Use ErrorMessages() to get an array of the messages only (no codes attached).
func (e *UserError) Error() string {
	return e.errorMessage(true)
}

// FriendlyError will return a string of one or all user errors.
//
// FriendlyError is equivalent to Error() except it doesn't include error codes in the message.
//
// If there is only one user error it will be represented as a single string.
//
//	Example:
//	   Email address is required
//
// If there are more than one user error (e.g. multiple validation errors)
// then a multi line string resembling a list of errors will be returned.
//
//	Example:
//	   - First name is required
//	   - Last name is required
//	   - Invalid email address
//
// Use Error() to compute the same string with error codes attached.
//
// Use ErrorMessages() to get an array of the messages only (no codes attached).
func (e *UserError) FriendlyError() string {
	return e.errorMessage(false)
}

// Errors returns a map of error codes and messages.
func (e *UserError) Errors() map[string]string {
	return e.errors
}

// ErrorMessages returns an array of error messages only.
func (e *UserError) ErrorMessages() []string {
	messages := make([]string, len(e.codes))
	for i, k := range e.codes {
		messages[i] = e.errors[k]
	}
	return messages
}

// User creates a new UserError fault.
func User(code string, msg string) *UserError {
	return &UserError{
		errors: map[string]string{
			code: msg,
		},
		codes: []string{code},
	}
}

// Userf creates a new UserError fault.
func Userf(code string, format string, a ...interface{}) *UserError {
	return User(code, fmt.Sprintf(format, a...))

}

// ------
// System Error
// ------

const (
	padding = "   "
)

// SystemError represents an error that was caused by an internal fault.
// A system error is typically an error which can only be handled by the application
// itself or would typically result in a 5xx status code in a web application context.
//
// Examples:
// - error connecting to a database
// - error reading from an IO stream
// - unexpected error from making a HTTP call
// - etc.
type SystemError struct {
	err   error
	msgs  []string
	stack string
}

// Error returns the error message.
func (e *SystemError) Error() string {
	pad := ""
	sb := strings.Builder{}
	lastIndex := len(e.msgs) - 1
	for i := lastIndex; i >= 0; i-- {
		if i < lastIndex {
			sb.WriteString(fmt.Sprintf("\n%s", pad))
		}
		sb.WriteString(e.msgs[i])
		pad = pad + padding
	}
	return sb.String()
}

// StackTrace returns the error message including the stack trace.
func (e *SystemError) StackTrace() string {
	return e.stack
}

// String returns the error message and stack trace.
func (e *SystemError) String() string {
	return fmt.Sprintf("%s\n%s", e.Error(), e.StackTrace())
}

// Unwrap returns the original underlying error.
func (e *SystemError) Unwrap() error {
	return e.err
}

// Format implements the fmt.Formatter interface.
// Implementation inspired by:
// https://github.com/pkg/errors/blob/5dd12d0cfe7f152f80558d591504ce685299311e/errors.go#L165
func (e *SystemError) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			fmt.Fprintf(s, "%s", e.String())
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Error())
	case 'q':
		fmt.Fprintf(s, "%q", e.Error())
	}
}

// System creates a new SystemError fault whilst preserving the stack trace.
func System(msg string) *SystemError {
	return &SystemError{
		err:   errors.New(msg),
		msgs:  []string{msg},
		stack: stack.Capture().String(),
	}
}

// Systemf creates a new SystemError fault whilst preserving the stack trace.
func Systemf(format string, a ...interface{}) *SystemError {
	return System(fmt.Sprintf(format, a...))
}

// SystemWrap creates a new SystemError fault, wrapping an
// existing error and preserving the entire stack trace.
func SystemWrap(err error, msg string) *SystemError {
	var msgs []string

	// nolint: errorlint // Don't want to check the entire chain, just outer most error:
	if sysErr, ok := err.(*SystemError); ok {
		msgs = append(sysErr.msgs, msg)
	} else {
		msgs = []string{err.Error(), msg}
	}

	return &SystemError{
		err:   fmt.Errorf("%s\n%s%w", msg, padding, err),
		msgs:  msgs,
		stack: stack.Capture().String(),
	}
}

// SystemWrapf creates a new SystemError fault, wrapping an
// existing error and preserving the entire stack trace.
func SystemWrapf(
	err error,
	format string,
	a ...interface{}) *SystemError {
	return SystemWrap(err, fmt.Sprintf(format, a...))
}

// As is similar, but a slightly different take on the errors.As function.
// Rather than matching on an interface or type it matches on a generic predicate function.
// This has the benefit that it can be applied with functions which return private/internal interfaces or types.
// For example, it can be used with the status.FromError function from the google.golang.org/grpc package.
func As[T any](
	err error,
	predicate func(error) (T, bool),
) (T, bool) {
	var zeroValue T
	for err != nil {
		if t, ok := predicate(err); ok {
			return t, true
		}
		err = errors.Unwrap(err)
	}
	return zeroValue, false
}
