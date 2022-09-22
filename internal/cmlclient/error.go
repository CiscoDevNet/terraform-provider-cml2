package cmlclient

import "errors"

var (
	ErrSystemNotReady  = errors.New("system not ready")
	ErrElementNotFound = errors.New("element not found")
)
