package yoitsu

import (
	"go/ast"
	"strings"
	"unicode"
)

type GeneratedType interface {
	Cleanup() (GeneratedType, error)
	IsComplexObject() bool
	Merge(other GeneratedType) (GeneratedType, error)

	Type() string
	// SameType may modify the struct if forgiving is true to force equivalence if possible
	// See specific implementations for details
	SameType(other GeneratedType, forgiving bool) bool

	Imports() []string
	Representation() []ast.Decl

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
