package yoitsu

import (
	"go/ast"
	"strings"
	"unicode"
)

// GeneratedType is our JsonObject abstraction with type info
type GeneratedType interface {
	// Cleanup traverses the type and tries fitting struts in Maps. See ValidIdFunc for how to control this
	Cleanup() (GeneratedType, error)
	// IsComplexObject returns true if the GeneratedType is not a NativeType
	IsComplexObject() bool
	// Merge returns the result of the merge, this GeneratedType may also have been modified
	Merge(other GeneratedType) (GeneratedType, error)

	// Type returns the string to use while constructing go-ast
	Type() string
	// UnderLyingType may return the type itself, if there is no underlying type
	UnderLyingType() GeneratedType
	// SameType may modify the struct if forgiving is true to force equivalence if possible
	// See specific implementations for details
	SameType(other GeneratedType, forgiving bool) bool

	// Imports returns on unsorted slice of import statements needed to construct this type
	Imports() []string
	// Representation returns the ast.Decl that needs to be included in the ast.File for this type
	Representation() []ast.Decl

	// Copy returns a new, but equivalent, GeneratedType instance
	Copy() GeneratedType
}

func toSafeGoName(name string) string {
	// Ensure Field is a valid Name
	if len(name) > 0 && !unicode.IsLetter(rune(name[0])) && !strings.HasPrefix(name, "[]") {
		name = "F" + name
	}

	if len(name) > 0 {
		runes := []rune(name)
		runes[0] = unicode.ToUpper(runes[0])
		name = string(runes)
	}

	var camelCaseName string
	for i, r := range name {
		if r == '_' && i+1 < len(name) {
			continue
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			camelCaseName += "_"
			continue
		}

		if i > 0 && name[i-1] == '_' {
			camelCaseName += string(unicode.ToUpper(r))
		} else {
			camelCaseName += string(r)
		}
	}

	return camelCaseName
}
