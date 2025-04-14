package yoitsu

import "errors"

var (
	ErrNoData                  = errors.New("no data to parse types from")
	ErrCantMergeDifferentTypes = errors.New("can't merge different types")
	ErrUnknownType             = errors.New("unknown type")
	ErrCannotRegisterForType   = errors.New("cannot register for type")
	ErrSrcIsNotLoadAble        = errors.New("src is not loadable")
)
