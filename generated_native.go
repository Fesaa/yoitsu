package yoitsu

import (
	"fmt"
	"go/ast"
)

var (
	StringType    = NewNativeType("string", "")
	Float64Type   = NewNativeType("float64", "")
	BoolType      = NewNativeType("bool", "")
	InterfaceType = NewNativeType("interface{}", "")
	TimeType      = NewNativeType("time.Time", "time")
)

func NewNativeType(t string, i string) GeneratedType {
	return &nativeType{
		_type:   t,
		_import: i,
	}
}

// nativeType represents an existing go type
type nativeType struct {
	_type   string
	_import string
}

func (g *nativeType) Copy() GeneratedType {
	return NewNativeType(g._type, g._import)
}

func (g *nativeType) IsComplexObject() bool {
	return false
}

func (g *nativeType) Type() string {
	return g._type
}

func (g *nativeType) Merge(other GeneratedType) (GeneratedType, error) {
	if g.Type() == InterfaceType.Type() {
		return other, nil
	}

	if _, ok := other.(*nativeType); !ok {
		return nil, fmt.Errorf("nativeType %w", ErrCantMergeDifferentTypes)
	}

	if other.Type() == InterfaceType.Type() {
		return g, nil
	}

	if !g.SameType(other, false) {
		return nil, fmt.Errorf("nativeType %w", ErrCantMergeDifferentTypes)
	}

	return g, nil
}

func (g *nativeType) SameType(other GeneratedType, forgiving bool) bool {
	return g.Type() == other.Type()
}

func (g *nativeType) Imports() []string {
	if g._import != "" {
		return []string{g._import}
	}
	return nil
}

func (g *nativeType) Cleanup() (GeneratedType, error) {
	return g, nil
}

func (g *nativeType) Representation() []ast.Decl {
	panic("native go type do not need a separate representation")
}
