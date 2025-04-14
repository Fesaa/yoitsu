package yoitsu

import "time"

type NativeTypeParser func(JsonObject) (GeneratedType, bool)

// TimeParser is an example implementation of a NativeTypeParser for time.Time
//
// Parser.RegisterNativeType(StringType, TimeParser)
func TimeParser(obj JsonObject) (GeneratedType, bool) {
	s, ok := obj.(string)
	if !ok {
		return nil, false
	}

	_, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return TimeType, true
	}

	return nil, false
}
