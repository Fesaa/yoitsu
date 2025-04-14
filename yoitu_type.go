package yoitsu

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
	"unicode"
)

type GeneratedType interface {
	Name() string
	SetName(string)
	// SameType returns true if all fields are of the same type, their names do not need to match
	// And their union contains the same elements as each subset on her own
	SameType(other GeneratedType) bool
	JsonType() JsonType
	Imports() []string
	// SameJsonType return true if the underlying JsonType is the same. This is a simple name check
	SameJsonType(other GeneratedType) bool
	IsComplexObject() bool
	Merge(other GeneratedType) error
	Representation() []ast.Decl
}

func generatedSimpleObject(name string, t JsonType) GeneratedType {
	return &generatedType{
		jsonType: t,
		name:     name,
	}
}

func generatedArray(name string, t JsonType) GeneratedType {
	return &generatedArrayType{
		generatedType: generatedType{
			jsonType: t,
			name:     name,
		},
	}
}

type generatedMapType struct {
	generatedType
}

func (gt *generatedMapType) JsonType() JsonType {
	return JsonMapType{
		gt.jsonType,
	}
}

func (gt *generatedArrayType) generatedMapType(o GeneratedType) bool {
	return gt.JsonType().TypeName() == o.JsonType().TypeName()
}

func (gt *generatedMapType) SameType(o GeneratedType) bool {
	if !gt.SameJsonType(o) {
		return false
	}

	// JsonType is the same so must be an array
	other := o.(*generatedMapType)
	return gt.generatedType.SameType(&other.generatedType)
}

func (gt *generatedMapType) Merge(o GeneratedType) error {
	other, ok := o.(*generatedMapType)
	if !ok {
		return ErrCantMergeDifferentTypes
	}

	// Overwrite generic types, or ignore them
	if gt.jsonType.TypeName() == JsonInterface.TypeName() {
		gt.generatedType = other.generatedType
		return nil
	}

	if other.JsonType().TypeName() == JsonInterface.TypeName() {
		return nil
	}

	return gt.generatedType.Merge(&other.generatedType)
}

type generatedArrayType struct {
	generatedType
}

func (gt *generatedArrayType) JsonType() JsonType {
	return JsonArray{
		gt.jsonType,
	}
}

func (gt *generatedArrayType) SameJsonType(o GeneratedType) bool {
	return gt.JsonType().TypeName() == o.JsonType().TypeName()
}

func (gt *generatedArrayType) SameType(o GeneratedType) bool {
	if !gt.SameJsonType(o) {
		gt.SameJsonType(o)
		return false
	}

	// JsonType is the same so must be an array
	other := o.(*generatedArrayType)
	return gt.generatedType.SameType(&other.generatedType)
}

func (gt *generatedArrayType) Merge(o GeneratedType) error {
	other, ok := o.(*generatedArrayType)
	if !ok {
		return ErrCantMergeDifferentTypes
	}

	// Overwrite generic types, or ignore them
	if gt.jsonType.TypeName() == JsonInterface.TypeName() {
		gt.generatedType = other.generatedType
		return nil
	}

	if other.JsonType().TypeName() == JsonInterface.TypeName() {
		return nil
	}

	return gt.generatedType.Merge(&other.generatedType)
}

type GeneratedTypeMap = map[string]GeneratedType

type generatedType struct {
	jsonType JsonType
	imports  []string
	types    GeneratedTypeMap
	name     string
}

func (gt *generatedType) Name() string {
	return toSafeGoName(gt.name)
}

func (gt *generatedType) SetName(name string) {
	gt.name = name
}

func (gt *generatedType) SameType(o GeneratedType) bool {
	other, ok := o.(*generatedType)
	if !ok {
		return false
	}

	if len(gt.types) != len(other.types) {
		return false
	}

	for name, gType := range gt.types {
		otherGType, ok := other.types[name]
		if !ok {
			return false
		}

		if !gType.SameType(otherGType) {
			return false
		}

	}

	return true
}

func (gt *generatedType) SameJsonType(other GeneratedType) bool {
	if gt.jsonType.TypeName() == JsonInterface.TypeName() {
		return true
	}
	if other.JsonType().TypeName() == JsonInterface.TypeName() {
		return true
	}
	return gt.JsonType().TypeName() == other.JsonType().TypeName()
}

func (gt *generatedType) JsonType() JsonType {
	return gt.jsonType
}

func (gt *generatedType) Imports() []string {
	var imports []string
	for _, i := range gt.imports {
		imports = append(imports, i)
	}

	for _, gType := range gt.types {
		imports = append(imports, gType.Imports()...)
	}

	return imports
}

func (gt *generatedType) IsComplexObject() bool {
	_, object := gt.jsonType.(JsonObject)
	_, array := gt.jsonType.(JsonArray)
	_, jsonMap := gt.jsonType.(JsonMapType)
	return object || array || jsonMap
}

func (gt *generatedType) Merge(o GeneratedType) error {
	other, ok := o.(*generatedType)
	if !ok {
		return ErrCantMergeDifferentTypes
	}

	if _, ok = gt.JsonType().(JsonPrimitive); ok {
		return nil
	}

	for name, otherGType := range other.types {
		gType, ok := gt.types[name]
		if !ok {
			gt.types[name] = otherGType
			continue
		}

		if err := gType.Merge(otherGType); err != nil {
			return err
		}

		gt.types[name] = gType
	}

	return nil
}

func toSafeGoName(name string) string {
	// Ensure Field is a valid name
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
		if i > 0 && name[i-1] == '_' {
			camelCaseName += string(unicode.ToUpper(r))
		} else {
			camelCaseName += string(r)
		}
	}

	return camelCaseName
}

func (gt *generatedType) Representation() []ast.Decl {
	fieldList := ast.FieldList{}

	structDecl := ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(toSafeGoName(gt.jsonType.TypeName())),
				Type: &ast.StructType{
					Fields: &fieldList,
				},
			},
		},
	}

	newTypes := []ast.Decl{&structDecl}

	for name, g := range gt.types {
		typeName := g.JsonType().TypeName()
		if g.IsComplexObject() {
			typeName = toSafeGoName(g.JsonType().TypeName())
		}

		field := &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(toSafeGoName(name))},
			Type:  ast.NewIdent(typeName),
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("`json:\"%s\"`", name),
			},
		}

		fieldList.List = append(fieldList.List, field)

		if g.IsComplexObject() {
			newTypes = append(newTypes, g.Representation()...)
		}
	}

	return newTypes
}
