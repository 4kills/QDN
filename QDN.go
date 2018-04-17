package qdn

import (
	"errors"
	"reflect"
	"strconv"
)

// Raw byte array in the qdn format. Represents a go struct
type Raw []byte

// Marshal turns the given struct into a raw byte array
func Marshal(stru interface{}) (Raw, error) {
	if reflect.TypeOf(stru).Kind() != reflect.Struct {
		return Raw{}, errors.New("qdn.Marshal error: Provided parameter is no struct")
	}

	count := reflect.TypeOf(stru).NumField()
	if count < 1 {
		return Raw{}, errors.New("qdn.Marshal error: The struct does not contain any fields")
	}

	r := make(Raw, 0)
	r = append(r, setupRaw(reflect.TypeOf(stru))...)
	for i := 0; i < count; i++ {
		r = append(r, fieldNameToRaw(reflect.TypeOf(stru).Field(i).Name)...)
		r = append(r, fieldValToRaw(reflect.ValueOf(stru).Field(i), reflect.TypeOf(stru).Field(i).Type)...)
	}

	return append(r, byte('>')), nil
}

func fieldValToRaw(val reflect.Value, typ reflect.Type) []byte {
	if typ.Kind() == reflect.String {
		b := append([]byte{byte('"')}, []byte(val.String())...)
		return append(b, []byte{byte('"'), byte(',')}...)
	}

	return append([]byte(valToString(val, typ)), byte(','))
}

func valToString(val reflect.Value, typ reflect.Type) string {
	// TODO: Add more types here
	switch typ.Kind() {
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(val.Float(), byte('f'), -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(val.Float(), byte('f'), -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(val.Bool())
	default:
		return "type not found"
	}
}

func fieldNameToRaw(s string) []byte {
	return append([]byte(s), byte('='))
}

func setupRaw(struName reflect.Type) []byte {
	return append([]byte(struName.Name()), byte('<'))
}
