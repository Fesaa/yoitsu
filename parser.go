package yoitsu

import "fmt"

type (
	JsonObject = interface{}
	JsonMap    = map[string]JsonObject
	JsonArray  = []JsonObject
)

type Parser struct {
	yoitsu *Yoitsu
}

func NewParser(yoitsu *Yoitsu) *Parser {
	return &Parser{
		yoitsu: yoitsu,
	}
}

func (p *Parser) ParseRoot(name string, root JsonObject) (GeneratedType, error) {
	if root == nil {
		return nil, ErrNoData
	}

	gType, err := p.Parse(name, root)
	if err != nil {
		return nil, err
	}

	gType, err = gType.Cleanup()
	if err != nil {
		return nil, err
	}

	return gType, nil
}

func (p *Parser) Parse(name string, s JsonObject) (GeneratedType, error) {
	switch s.(type) {
	case JsonArray:
		return p.ParseArray(name, s.(JsonArray))
	case JsonMap:
		return p.ParseObject(name, s.(JsonMap))
	case JsonObject:
		return p.ParseNative(s.(JsonObject))
	case nil:
		return InterfaceType, nil
	}

	return nil, fmt.Errorf("%w: can't parse type %T", ErrUnknownType, s)
}

func (p *Parser) ParseArray(name string, array JsonArray) (GeneratedType, error) {
	var arrayType GeneratedType

	if len(array) == 0 {
		return &SliceType{InterfaceType}, nil
	}

	for _, v := range array {
		gType, err := p.Parse(SliceNameFormatter(name), v)
		if err != nil {
			return nil, err
		}

		if arrayType == nil {
			arrayType = gType
			continue
		}

		arrayType, err = arrayType.Merge(gType)
		if err != nil {
			return nil, err
		}
	}

	return &SliceType{p.yoitsu.universe.FindType(arrayType)}, nil
}

func (p *Parser) ParseObject(name string, obj JsonMap) (GeneratedType, error) {
	if len(obj) == 0 {
		return nil, ErrNoData
	}

	st := StructType{
		Name:   toSafeGoName(name),
		Fields: make(map[string]*StructField),
	}

	for jsonName, jsonObject := range obj {
		gType, err := p.Parse(name+jsonName, jsonObject)
		if err != nil {
			return nil, err
		}

		if stGType, ok := gType.(*StructType); ok {
			stGType.tag = jsonName
		}

		st.Fields[jsonName] = &StructField{
			Type: gType,
			Tag:  jsonName,
		}
	}

	return p.yoitsu.universe.FindType(&st), nil
}

func (p *Parser) ParseNative(obj JsonObject) (GeneratedType, error) {
	switch obj.(type) {
	case float64:
		return Float64Type, nil
	case string:
		return StringType, nil
	case bool:
		return BoolType, nil
	case nil:
		return InterfaceType, nil
	}

	return nil, fmt.Errorf("%w: can't parse type %T", ErrUnknownType, obj)
}
