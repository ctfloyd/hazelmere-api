package database

import "errors"

var ErrGeneric = errors.New("generic database service_error")
var ErrNotFound = errors.New("not found")
