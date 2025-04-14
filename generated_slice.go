package yoitsu

import (
	"fmt"
	"go/ast"
	"strings"
)

var SliceNameFormatter = sliceNameFormatter

type SliceType struct {
	SliceType GeneratedType
}

func (s *SliceType) Copy() GeneratedType {
	return &SliceType{
		SliceType: s.SliceType.Copy(),
	}
}

func (s *SliceType) UnderLyingType() GeneratedType {
	return s.SliceType
}

func (s *SliceType) IsComplexObject() bool {
	return true
}

func (s *SliceType) Merge(other GeneratedType) (GeneratedType, error) {
	otherSlice, ok := other.(*SliceType)
	if !ok {
		return nil, fmt.Errorf("SliceType %w", ErrCantMergeDifferentTypes)
	}

	newType, err := s.SliceType.Merge(otherSlice.SliceType)
	if err != nil {
		return nil, err
	}
	s.SliceType = newType
	return s, nil
}

func (s *SliceType) Type() string {
	return fmt.Sprintf("[]%s", s.SliceType.Type())
}

func (s *SliceType) SameType(other GeneratedType, forgiving bool) bool {
	if sliceType, ok := other.(*SliceType); ok {
		return s.SliceType.SameType(sliceType.SliceType, forgiving)
	}
	return false
}

func (s *SliceType) Cleanup() (GeneratedType, error) {
	sliceType, err := s.SliceType.Cleanup()
	if err != nil {
		return nil, err
	}
	s.SliceType = sliceType
	return s, nil
}

func (s *SliceType) Imports() []string {
	return s.SliceType.Imports()
}

func (s *SliceType) Representation() []ast.Decl {
	if s.SliceType.IsComplexObject() {
		return s.SliceType.Representation()
	}
	return nil
}

func sliceNameFormatter(s string) string {
	suffixes := []string{"item"}

	for _, suffix := range suffixes {
		l := len(suffix)
		if len(s) < l {
			continue
		}

		suff := s[len(s)-l:]

		if strings.ToLower(suff) == suffix {
			return s
		}
	}

	return s + "Item"
}

/*func sliceNameFormatter(s string) string {
	suffixes := []string{"list", "array", "slice"}

	for _, suffix := range suffixes {
		l := len(suffix)
		if len(s) < l {
			continue
		}

		suff := s[len(s)-l:]

		if strings.ToLower(suff) == suffix {
			return s[:len(s)-l]
		}
	}

	return s
}*/
