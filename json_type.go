package yoitsu

import "fmt"

type JsonMap = map[string]interface{}

type JsonType interface {
	TypeName() string
}

type JsonObject struct {
	inner JsonPrimitive
}

func (j JsonObject) TypeName() string {
	return j.inner.TypeName()
}

type JsonArray struct {
	_type JsonType
}

func (j JsonArray) TypeName() string {
	return fmt.Sprintf("[]%s", j._type.TypeName())
}

type JsonPrimitive struct {
	_type string
}

func (j JsonPrimitive) TypeName() string {
	return j._type
}

var (
	JsonString    JsonType = JsonPrimitive{"string"}
	JsonInt       JsonType = JsonPrimitive{"int"}
	JsonInt8      JsonType = JsonPrimitive{"int8"}
	JsonInt16     JsonType = JsonPrimitive{"int16"}
	JsonInt32     JsonType = JsonPrimitive{"int32"}
	JsonInt64     JsonType = JsonPrimitive{"int64"}
	JsonUint      JsonType = JsonPrimitive{"uint"}
	JsonUint8     JsonType = JsonPrimitive{"uint8"}
	JsonUint16    JsonType = JsonPrimitive{"uint16"}
	JsonUint32    JsonType = JsonPrimitive{"uint32"}
	JsonUint64    JsonType = JsonPrimitive{"uint64"}
	JsonFloat64   JsonType = JsonPrimitive{"float64"}
	JsonFloat32   JsonType = JsonPrimitive{"float32"}
	JsonBool      JsonType = JsonPrimitive{"bool"}
	JsonInterface JsonType = JsonPrimitive{"interface{}"}
)
