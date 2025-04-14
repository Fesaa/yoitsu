package yoitsu

import "slices"

type Universe interface {
	// FindType returns the passed type is no type is found
	FindType(GeneratedType) GeneratedType
	AddType(GeneratedType)
}

// EmptyUniverse returns a universe that always does a no-op when called upon
func EmptyUniverse() Universe {
	return emptyUniverse{}
}

func NewUniverse() Universe {
	return &universe{
		_types: make([]GeneratedType, 0),
	}
}

type emptyUniverse struct{}

func (e emptyUniverse) FindType(generatedType GeneratedType) GeneratedType {
	return generatedType
}

func (e emptyUniverse) AddType(generatedType GeneratedType) {
}

type universe struct {
	_types []GeneratedType
}

func (u *universe) FindType(generatedType GeneratedType) GeneratedType {
	for _, t := range u._types {
		if t.SameType(generatedType) {
			return t
		}
	}
	return generatedType
}

func (u *universe) AddType(generatedType GeneratedType) {
	u._types = append(u._types, generatedType)
}

func WithField(name string, jsonType JsonType) Option[*generatedType] {
	return func(gt *generatedType) {
		gt.types[name] = generatedSimpleObject(name, jsonType)
	}
}

func WithImport(i string) Option[*generatedType] {
	return func(gt *generatedType) {
		if !slices.Contains(gt.imports, i) {
			gt.imports = append(gt.imports, i)
		}
	}
}

func ExistingTypeWrapper(name string, opts ...Option[*generatedType]) GeneratedType {
	gt := &generatedType{
		jsonType: JsonPrimitive{name},
		imports:  make([]string, 0),
		types:    make(GeneratedTypeMap),
	}

	for _, opt := range opts {
		opt(gt)
	}

	return gt
}
