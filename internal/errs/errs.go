package errs

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrLoginTaken         = errors.New("login already taken")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("token is invalid")
)
