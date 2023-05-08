package simple_http

import (
	"encoding/json"
	"reflect"
	"strconv"
)

func ToString(src interface{}) string {
	switch v := src.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	bytes, _ := json.Marshal(src)
	return string(bytes)
}

// Map2MapString
//
//	@param src
//	@return map
func Map2MapString(src map[string]interface{}) map[string]string {
	retData := map[string]string{}
	for k, v := range src {
		retData[k] = ToString(v)
	}
	return retData
}

// ToInterfaceMap
//
//	@param src
//	@return map
func ToInterfaceMap(src map[string]string) map[string]interface{} {
	retData := map[string]interface{}{}
	for k, v := range src {
		retData[k] = v
	}
	return retData
}
