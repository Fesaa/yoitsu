package yoitsu

import (
	"fmt"
	"go/ast"
	"go/token"
)

func ParseTypes(name string, data []interface{}, universe Universe) (GeneratedType, error) {
	types := make([]GeneratedType, len(data))

	if len(data) == 0 {
		return nil, nil
	}

	for i, v := range data {
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

		types[i] = gType
	}

	if len(types) == 0 {
		return &generatedType{}, fmt.Errorf("%s: no types found", name)
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
	}

	return nil, fmt.Errorf("unknown type found in json %T", value)
}

func (gt *generatedType) Representation() []*ast.GenDecl {
	structDecl := ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: ast.NewIdent(gt.JsonType().TypeName()),
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{},
					},
				},
			},
		},
	}

	innerTypes := []*ast.GenDecl{&structDecl}

	for name, g := range gt.types {
		l := structDecl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List

		field := &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(name)},
			Type:  ast.NewIdent(g.JsonType().TypeName()),
			Tag: &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("`json:\"%s\"`", name),
			},
		}

		structDecl.Specs[0].(*ast.TypeSpec).Type.(*ast.StructType).Fields.List = append(l, field)

		if g.IsComplexObject() {
			innerTypes = append(innerTypes, g.Representation()...)
		}

	}

	return innerTypes
}
