package yoitsu

import (
	"fmt"
	"go/ast"
)

type MapType struct {
	ValueType GeneratedType
}

func (m *MapType) Copy() GeneratedType {
	return &MapType{
		ValueType: m.ValueType.Copy(),
	}
}

func (m *MapType) UnderLyingType() GeneratedType {
	return m.ValueType
}

func (m *MapType) IsComplexObject() bool {
	return true
}

func (m *MapType) Merge(other GeneratedType) (GeneratedType, error) {
	mType, ok := other.(*MapType)
	if !ok {
		return nil, fmt.Errorf("MapType %w %T", ErrCantMergeDifferentTypes, other)
	}

	newType, err := m.ValueType.Merge(mType.ValueType)
	if err != nil {
		return nil, err
	}

	m.ValueType = newType
	return m, nil
}

func (m *MapType) Type() string {
	return fmt.Sprintf("map[string]%s", m.ValueType.Type())
}

func (m *MapType) SameType(other GeneratedType, forgiving bool) bool {
	if mType, ok := other.(*MapType); ok {
		return m.ValueType.SameType(mType.ValueType, forgiving)
	}
	return false
}

func (m *MapType) Imports() []string {
	return m.ValueType.Imports()
}

func (m *MapType) Cleanup() (GeneratedType, error) {
	panic("maps can't be cleaned up")
}

func (m *MapType) Representation() []ast.Decl {
	if m.ValueType.IsComplexObject() {
		return m.ValueType.Representation()
	}
	return nil
}
