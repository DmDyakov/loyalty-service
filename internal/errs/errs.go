package errs

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid login or password")
	ErrLoginTaken         = errors.New("login already taken")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("token is invalid")
)

var (
	ErrOrderUploadedByAnother = errors.New("order has been uploaded by another user")
	ErrOrderAlreadyExists     = errors.New("order has been already uploaded")
	ErrInvalidOrderNumber     = errors.New("invalid order number")
	ErrOrderStatusesRequired  = errors.New("at least one order status is required")
)

var (
	ErrUnsupportedLimit = errors.New("unsupported limit")
)
