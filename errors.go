package yoitsu

import "errors"

var (
	ErrNoData                  = errors.New("no data to parse types from")
	ErrCantMergeDifferentTypes = errors.New("can't merge different types")
	ErrAccessorsNotGenerated   = errors.New("accessors are not generated for non arrays")
	ErrUnknownType             = errors.New("unknown type")
)
