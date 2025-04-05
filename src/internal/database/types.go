package database

import "errors"

var ErrGeneric = errors.New("generic database error")
var ErrNotFound = errors.New("not found")
