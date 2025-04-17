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

func NewNativeType(typeName string, importLine string) GeneratedType {
	return &NativeType{
		_type:   typeName,
		_import: importLine,
	}
}

// NativeType represents an existing go type
//
// See Parser.RegisterNativeType for how to expand the current list
type NativeType struct {
	_type   string
	_import string
}

func (g *NativeType) UnderLyingType() GeneratedType {
	return g
}

func (g *NativeType) Copy() GeneratedType {
	return NewNativeType(g._type, g._import)
}

func (g *NativeType) IsComplexObject() bool {
	return false
}

func (g *NativeType) Type() string {
	return g._type
}

func (g *NativeType) Merge(other GeneratedType) (GeneratedType, error) {
	if g.Type() == InterfaceType.Type() {
		return other, nil
	}

	if _, ok := other.(*NativeType); !ok {
		return nil, fmt.Errorf("nativeType %w", ErrCantMergeDifferentTypes)
	}

	if other.Type() == InterfaceType.Type() {
		return g, nil
	}

	if !g.SameType(other, false) {
		return nil, fmt.Errorf("NativeType %w", ErrCantMergeDifferentTypes)
	}

	return g, nil
}

func (g *NativeType) SameType(other GeneratedType, forgiving bool) bool {
	return g.Type() == other.Type()
}

func (g *NativeType) Imports() []string {
	if g._import != "" {
		return []string{g._import}
	}
	return nil
}

func (g *NativeType) Cleanup() (GeneratedType, error) {
	return g, nil
}

func (g *NativeType) Representation() []ast.Decl {
	panic("native go type do not need a separate representation")
}
