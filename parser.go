package yoitsu

import "fmt"

type (
	JsonObject = interface{}
	JsonMap    = map[string]JsonObject
	JsonArray  = []JsonObject
)

// Parser is used to parse the JsonObject into a GeneratedType
type Parser struct {
	yoitsu *Yoitsu

	stringParsers []NativeTypeParser
	floatParsers  []NativeTypeParser
}

func NewParser(yoitsu *Yoitsu) *Parser {
	return &Parser{
		yoitsu: yoitsu,
	}
}

// RegisterNativeType registers a new NativeType to be used when parsing a string or float64
func (p *Parser) RegisterNativeType(parent GeneratedType, f NativeTypeParser) error {
	switch parent.Type() {
	case StringType.Type():
		p.stringParsers = append(p.stringParsers, f)
		return nil
	case Float64Type.Type():
		p.floatParsers = append(p.floatParsers, f)
		return nil
	}

	return fmt.Errorf("%w: %s", ErrCannotRegisterForType, parent.Type())

}

// ParseRoot calls Parse and then GeneratedType.Cleanup
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

// Parse recursively traverses the JsonObject to construct the GeneratedType
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
		return p.parseFloat64(obj.(float64))
	case string:
		return p.parseString(obj.(string))
	case bool:
		return BoolType, nil
	case nil:
		return InterfaceType, nil
	}

	return nil, fmt.Errorf("%w: can't parse type %T", ErrUnknownType, obj)
}

func (p *Parser) parseString(s string) (GeneratedType, error) {
	for _, parser := range p.stringParsers {
		if gType, ok := parser(s); ok {
			return gType, nil
		}
	}

	return StringType, nil
}

func (p *Parser) parseFloat64(f float64) (GeneratedType, error) {
	for _, parser := range p.floatParsers {
		if gType, ok := parser(f); ok {
			return gType, nil
		}
	}

	// Not every number fits inside a float64, even tho go sets the type as such ...
	// We're generating code, so don't know what the number *CAN* be. Lets be safe
	return &StructType{
		Name:   "json.Number",
		Import: "encoding/json",
	}, nil
}
