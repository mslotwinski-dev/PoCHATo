package main

import "errors"

var (
	ErrUnauthorized   = errors.New("unauthorized")
	ErrUserNotFound   = errors.New("user not found")
	ErrInvalidRequest = errors.New("invalid request")
)
