package yoitsu

// Universe holds the GeneratedType available for use while parsing
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
		if t.SameType(generatedType, false) {
			return t.Copy()
		}
	}
	return generatedType
}

func (u *universe) AddType(generatedType GeneratedType) {
	u._types = append(u._types, generatedType)
}
