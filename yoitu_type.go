package yoitsu

import (
	"fmt"
	"go/ast"
	"go/token"
	"unicode"
)

func ParseTypes(name string, data []interface{}, universe Universe) (GeneratedType, error) {
	types := make([]GeneratedType, 0)

	if len(data) == 0 {
		return nil, nil
	}

	for _, v := range data {
		var gType GeneratedType
		var err error

		switch v.(type) {
		case JsonMap:
			gType, err = ParseType(name, v.(JsonMap), universe)
		case []interface{}:
			gType, err = ParseTypes(name, v.([]interface{}), universe)
		default:
			gType, err = parse(name, v, universe)
		}

		if err != nil {
			return nil, err
		}

		if gType == nil {
			continue
		}

		types = append(types, gType)
	}

	if len(types) == 0 {
		return &generatedType{}, ErrorNoData
	}

	gt := types[0]
	for _, t := range types[1:] {
		if err := gt.Merge(t); err != nil {
			return &generatedType{}, err
		}
	}

	return gt, nil
}

func ParseType(name string, data JsonMap, universe Universe) (GeneratedType, error) {
	gt := generatedType{
		jsonType: JsonObject{JsonPrimitive{name}},
		types:    make(GeneratedTypeMap),
	}

	if err := gt.parse(data, universe); err != nil {
		return &generatedType{}, err
	}
	return &gt, nil
}

func generatedSimpleObject(t JsonType) GeneratedType {
	return &generatedType{
		jsonType: t,
	}
}

func generatedArray(t JsonType) GeneratedType {
	return &generatedArrayType{
		generatedType: generatedType{
			jsonType: t,
		},
	}
}

type GeneratedType interface {
	// SameType returns true if all fields are of the same type, their names do not need to match
	// And their union contains the same elements as each subset on her own
	SameType(other GeneratedType) bool
	JsonType() JsonType
	Imports() []string
	// SameJsonType return true if the underlying JsonType is the same. This is a simple name check
	SameJsonType(other GeneratedType) bool
	IsComplexObject() bool
	Merge(other GeneratedType) error
	Representation() []*ast.GenDecl
}

type generatedArrayType struct {
	generatedType
}

func (gt *generatedArrayType) JsonType() JsonType {
	return JsonArray{
		gt.jsonType,
	}
}

func (gt *generatedArrayType) SameType(o GeneratedType) bool {
	if !gt.SameJsonType(o) {
		return false
	}

	// JsonType is the same so must be an array
	other := o.(*generatedArrayType)
	return gt.generatedType.SameType(&other.generatedType)
}

func (gt *generatedArrayType) Merge(o GeneratedType) error {
	other, ok := o.(*generatedArrayType)
	if !ok {
		return fmt.Errorf("cannot merge array with non array type")
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
}

func (gt *generatedType) SameType(o GeneratedType) bool {
	// JsonType is the same, so can't be an array
	other := o.(*generatedType)

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
	return gt.jsonType.TypeName() == other.JsonType().TypeName()
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
	return object || array
}

func (gt *generatedType) Merge(o GeneratedType) error {
	if !gt.SameJsonType(o) {
		return fmt.Errorf("cannot merge different jsonType types %s -> %s", gt.jsonType, o.JsonType())
	}
	other := o.(*generatedType)

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

func (gt *generatedType) parse(data JsonMap, universe Universe) error {
	for name, value := range data {
		gType, err := parse(name, value, universe)
		if err != nil {
			return err
		}

		gt.types[name] = gType
	}

	return nil
}

func parse(name string, value interface{}, universe Universe) (GeneratedType, error) {
	switch value.(type) {
	case string:
		return generatedSimpleObject(JsonString), nil
	case float64:
		return generatedSimpleObject(JsonFloat64), nil
	case float32:
		return generatedSimpleObject(JsonFloat32), nil
	case bool:
		return generatedSimpleObject(JsonBool), nil
	case int:
		return generatedSimpleObject(JsonInt), nil
	case uint:
		return generatedSimpleObject(JsonUint), nil
	case int64:
		return generatedSimpleObject(JsonInt64), nil
	case uint64:
		return generatedSimpleObject(JsonUint64), nil
	case int32:
		return generatedSimpleObject(JsonInt32), nil
	case uint32:
		return generatedSimpleObject(JsonUint32), nil
	case int16:
		return generatedSimpleObject(JsonInt16), nil
	case uint16:
		return generatedSimpleObject(JsonUint16), nil
	case int8:
		return generatedSimpleObject(JsonInt8), nil
	case uint8:
		return generatedSimpleObject(JsonUint8), nil
	case JsonMap:
		objectType, err := ParseType(name, value.(JsonMap), universe)
		if err != nil {
			return nil, err
		}
		return universe.FindType(objectType), nil
	case []interface{}:
		arrayType, err := ParseTypes(name, value.([]interface{}), universe)
		if err != nil {
			return nil, err
		}
		if arrayType != nil {
			arrayType = universe.FindType(arrayType)
			return &generatedArrayType{*arrayType.(*generatedType)}, nil
		} else {
			// The array contained no elements, we can't know what is inside it
			return generatedArray(JsonInterface), nil
		}
	case nil:
		return nil, nil
	}

	return nil, fmt.Errorf("unknown type found in json %T", value)
}

func toSafeGoName(name string) string {
	// Ensure Field is a valid name
	if len(name) > 0 && !unicode.IsLetter(rune(name[0])) {
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

func (gt *generatedType) Representation() []*ast.GenDecl {
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

	newTypes := []*ast.GenDecl{&structDecl}

	for name, g := range gt.types {
		field := &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(toSafeGoName(name))},
			Type:  ast.NewIdent(g.JsonType().TypeName()),
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
