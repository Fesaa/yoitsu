package yoitsu

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strconv"
	"strings"
)

// ValidIdFunc decided if a field name is an ID, if all field names are IDs, and all StructField.Type's are
// the same, the StructType is converted into a MapType
//
// See StructType.Cleanup for more info on semantics
var ValidIdFunc = defaultValidIdFunc

// ShouldConvertToMap decided if a StructType should be converted into a MapType
//
// See StructType.Cleanup for default behavior
var ShouldConvertToMap = defaultShouldConvertToMapFunc

// StructType represents a "smart" JsonMap
//
// # A StructType may be converted back to a MapType during cleanup, see ValidIdFunc to customize this behavior
//
// Add this type to your own (non-empty) Universe to have the Parser (re-)use your own types
type StructType struct {
	Name   string
	Import string
	Fields map[string]*StructField

	tag string
}

// StructField represents a field in a StructType
type StructField struct {
	Type GeneratedType
	Tag  string
}

func (s *StructType) UnderLyingType() GeneratedType {
	return s
}

func (s *StructType) Copy() GeneratedType {
	fields := make(map[string]*StructField)
	for k, v := range s.Fields {
		fields[k] = &StructField{
			Type: v.Type.Copy(),
			Tag:  v.Tag,
		}
	}

	return &StructType{
		Name:   s.Name,
		tag:    s.tag,
		Import: s.Import,
		Fields: fields,
	}
}

func (s *StructType) IsComplexObject() bool {
	return true
}

func (s *StructType) Merge(other GeneratedType) (GeneratedType, error) {
	if other.SameType(InterfaceType, false) {
		return s, nil
	}

	st, ok := other.(*StructType)
	if !ok {
		return nil, fmt.Errorf("StructType %w (%T)", ErrCantMergeDifferentTypes, other)
	}

	for tag, field := range st.Fields {
		existingField, ok := s.Fields[tag]
		if !ok {
			s.Fields[tag] = field
			continue
		}

		newFieldType, err := existingField.Type.Merge(field.Type)
		if err != nil {
			return nil, err
		}

		existingField.Type = newFieldType
		s.Fields[tag] = existingField
	}

	if st.Import == "" && s.Import != "" { // Reset import and take non-imported name if merged is a new type
		s.Name = st.Name
		s.Import = ""
	} else if st.Import == "" && s.Import == "" { // When merging, always use the "smallest" name to ensure predictable
		if s.Name >= st.Name {
			s.Name = st.Name
			s.tag = st.tag
		}
	} else if st.Import != "" {
		s.Import = st.Import
		s.Name = st.Name
	}

	return s, nil
}

func (s *StructType) Type() string {
	return s.Name
}

func (s *StructType) SameType(other GeneratedType, forgiving bool) bool {
	sOther, ok := other.(*StructType)
	if !ok {
		return false
	}

	for tag, field := range s.Fields {
		sField, ok := sOther.Fields[tag]
		if !ok {
			if !forgiving {
				return false
			}

			continue
		}

		if !field.Type.SameType(sField.Type, forgiving) {
			return false
		}
	}

	for tag, _ := range sOther.Fields {
		_, ok := s.Fields[tag]
		if ok {
			// Already checked in the loop above
			continue
		}

		if !forgiving {
			return false
		}
	}

	return true
}

// Cleanup will cleanup all fields, then check if this StructType could be a MapType.
//
// # The following conditions must be met
//
// - If not all fields are an Id, all fields must be GeneratedType.IsComplexObject
//
// - If there are less than two fields, the field must be GeneratedType.IsComplexObject
//
// - There must be at least one field
//
// - The type of the fields must not be of type InterfaceType
//
// Overwrite this behaviour by changing the ShouldConvertToMap function
func (s *StructType) Cleanup() (GeneratedType, error) {
	// Cleanup children
	for tag, field := range s.Fields {
		fType, err := field.Type.Cleanup()
		if err != nil {
			return nil, err
		}

		field.Type = fType
		s.Fields[tag] = field
	}

	// If all fields are the same type, convert into map
	var tracker GeneratedType
	allComplex := true
	allIds := true

	for _, field := range s.Fields {
		allComplex = allComplex && field.Type.IsComplexObject()
		allIds = allIds && ValidIdFunc(field.Tag)

		if tracker == nil {
			tracker = field.Type
			continue
		}

		if !tracker.SameType(field.Type, true) {
			return s, nil
		}
	}

	if tracker == nil {
		return s, nil
	}

	if !ShouldConvertToMap(s, allComplex, allIds, tracker) {
		return s, nil
	}

	// Merge all types to ensure we have all fields, and a predictable name
	for _, field := range s.Fields {
		var err error
		tracker, err = tracker.Merge(field.Type)
		if err != nil {
			return nil, err
		}
	}

	if st, ok := tracker.(*StructType); ok {
		if st.tag != "" {
			st.removeFromName(st.tag)
		} else {
			st.removeFromName(st.Name)
		}
	}

	return &MapType{
		ValueType: tracker,
	}, nil
}

func (s *StructType) removeFromName(str string) {
	if s.Import != "" {
		return
	}

	s.Name = strings.ReplaceAll(s.Name, str, "")

	for _, field := range s.Fields {
		if st, ok := field.Type.UnderLyingType().(*StructType); ok {
			st.removeFromName(str)
		}
	}
}

func (s *StructType) Imports() []string {
	var imports []string

	if s.Import != "" {
		return append(imports, s.Import)
	}

	for _, field := range s.Fields {
		for _, i := range field.Type.Imports() {
			if !slices.Contains(imports, i) {
				imports = append(imports, i)
			}
		}
	}

	return imports
}

func (s *StructType) Representation() []ast.Decl {
	if s.Import != "" {
		return nil
	}

	fieldList := ast.FieldList{}

	structDecl := ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(s.Type()),
				Type: &ast.StructType{
					Fields: &fieldList,
				},
			},
		},
	}

	newTypes := []ast.Decl{&structDecl}

	var tags []string
	for tag, _ := range s.Fields {
		tags = append(tags, tag)
	}
	slices.Sort(tags)

	for _, tag := range tags {
		field := s.Fields[tag]

		fieldList.List = append(fieldList.List, &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(toSafeGoName(field.Tag))},
			Type:  ast.NewIdent(field.Type.Type()),
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("`json:\"%s\"`", field.Tag),
			},
		})

		if field.Type.IsComplexObject() {
			newTypes = append(newTypes, field.Type.Representation()...)
		}
	}

	return newTypes
}

func defaultShouldConvertToMapFunc(s *StructType, allComplex, allIds bool, tracker GeneratedType) bool {
	if !allComplex && !allIds {
		return false
	}

	if len(s.Fields) < 2 && !allComplex {
		return false
	}

	if tracker == nil {
		return false
	}

	if tracker.SameType(InterfaceType, false) {
		return false
	}

	return true
}

func defaultValidIdFunc(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	return false
}
