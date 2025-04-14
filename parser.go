package yoitsu

import "fmt"

func ParseTypes(name string, data []interface{}, y *Yoitsu) (GeneratedType, error) {
	types := make([]GeneratedType, 0)

	if len(data) == 0 {
		return nil, nil
	}

	for _, v := range data {
		var gType GeneratedType
		var err error

		switch v.(type) {
		case JsonMap:
			gType, err = ParseType(name, v.(JsonMap), y)
		case []interface{}:
			gType, err = ParseTypes(name, v.([]interface{}), y)
		default:
			gType, err = parse(name, v, y)
		}

		if err != nil {
			return nil, err
		}

		if gType == nil {
			continue
		}

		types = append(types, y.universe.FindType(gType))
	}

	mergedType, err := MergeTypes(types)
	if err != nil {
		return nil, err
	}

	return &generatedArrayType{*mergedType.(*generatedType)}, nil
}

func ParseType(name string, data JsonMap, y *Yoitsu) (GeneratedType, error) {
	gt := generatedType{
		jsonType: JsonObject{name},
		types:    make(GeneratedTypeMap),
	}

	if err := gt.parse(data, y); err != nil {
		return &generatedType{}, err
	}
	gt.name = name
	return &gt, nil
}

func ParseTypeSmart(name string, data JsonMap, y *Yoitsu) (GeneratedType, error) {
	gt, err := ParseType(name, data, y)
	if err != nil {
		return &generatedType{}, err
	}

	gType := gt.(*generatedType)

	// All types must be the same
	var (
		t     GeneratedType
		cmpl  = true
		types []GeneratedType
	)

	for _, field := range gType.types {
		if t == nil {
			t = field
		}

		if !t.SameType(field) {
			return gt, err
		}

		cmpl = cmpl && field.IsComplexObject()
		types = append(types, field)
	}

	mergedGType, err := MergeTypes(types)
	if err != nil {
		return &generatedType{}, err
	}

	mergedType := *mergedGType.(*generatedType)
	// Update name
	mergedType.jsonType = JsonObject{name}
	mergedType.SetName(name)

	// Update root
	newRoot := make([]interface{}, len(data))
	for _, v := range data {
		newRoot = append(newRoot, v)
	}
	y.root = newRoot

	return &generatedMapType{mergedType}, nil
}

func MergeTypes(types []GeneratedType) (GeneratedType, error) {
	if len(types) == 0 {
		return &generatedType{}, ErrNoData
	}

	gt := types[0]
	for _, t := range types[1:] {
		if err := gt.Merge(t); err != nil {
			return &generatedType{}, err
		}
	}

	return gt, nil
}

func (gt *generatedType) parse(data JsonMap, y *Yoitsu) error {
	for name, value := range data {
		gType, err := parse(gt.JsonType().TypeName()+name, value, y)
		if err != nil {
			return err
		}

		gType.SetName(name)
		gt.types[name] = gType
	}

	return nil
}

func parse(name string, value interface{}, y *Yoitsu) (GeneratedType, error) {
	switch value.(type) {
	case string:
		return generatedSimpleObject(name, JsonString), nil
	case float64:
		return generatedSimpleObject(name, JsonFloat64), nil
	case bool:
		return generatedSimpleObject(name, JsonBool), nil
	case JsonMap:
		objectType, err := ParseType(name, value.(JsonMap), y)
		if err != nil {
			return nil, err
		}

		return y.universe.FindType(objectType), nil
	case []interface{}:
		arrayType, err := ParseTypes(name, value.([]interface{}), y)
		if err != nil {
			return nil, err
		}
		if arrayType != nil {
			at := arrayType.(*generatedArrayType)
			overWrite := y.universe.FindType(&at.generatedType)

			return &generatedArrayType{*overWrite.(*generatedType)}, nil
		} else {
			// The array contained no elements, we can't know what is inside it
			return generatedArray(name, JsonInterface), nil
		}
	case nil:
		return nil, nil
	}

	return nil, fmt.Errorf("unknown type found in json %T", value)
}
