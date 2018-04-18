package qdn

import (
	"bytes"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

// Raw byte array in the qdn format. Represents a go struct
type Raw []byte

// Format puts the raw byte data in a more readable state, for use in a text editor.
// Keep in mind that this uses system resources, so its not adviseable to use for network transmission
func (r *Raw) Format() error {
	// not implemented
	return nil
}

// Unmarshal fills a given interface with the corresponding qdn byte data.
func Unmarshal(stru *interface{}, data []byte) error {
	if reflect.TypeOf(stru).Kind() != reflect.Struct {
		return errors.New("qdn.Unmarshal error: Provided parameter is no struct pointer")
	}

	if !bytes.ContainsAny(data[:], reflect.TypeOf(stru).Name()) {
		return errors.New("qdn.Unmarshal error: Missmatch: data does not represent the struct")
	}

	count := reflect.TypeOf(stru).NumField()
	if count < 1 {
		return errors.New("qdn.Unmarshal error: The struct does not contain any fields")
	}

	var (
		at  int
		err error
	)
	for i := 0; i < count; i++ {
		at, err = strToVal(reflect.ValueOf(stru).Field(i), reflect.TypeOf(stru).Field(i).Type, data, at)
		if err != nil {
			return err
		}
	}

	return nil
}

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

func strToVal(val reflect.Value, typ reflect.Type, data []byte, at int) (int, error) {
	i := bytes.IndexRune(data[at:], '=')
	at = bytes.IndexRune(data[at:], ',')
	s := string(data[i+1 : at])

	switch typ.Kind() {
	case reflect.String:
		val.SetString(s)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return -1, errors.New("Conversion error: " + err.Error())
		}
		val.SetInt(n)
	case reflect.Float32:
		n, err := strconv.ParseFloat(s, 32)
		if err != nil {
			return -1, errors.New("Conversion error: " + err.Error())
		}
		val.SetFloat(n)
	case reflect.Float64:
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return -1, errors.New("Conversion error: " + err.Error())
		}
		val.SetFloat(n)
	case reflect.Bool:
		n, err := strconv.ParseBool(s)
		if err != nil {
			return -1, errors.New("Conversion error: " + err.Error())
		}
		val.SetBool(n)
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return -1, errors.New("Conversion error: " + err.Error())
		}
		val.SetUint(n)
	case reflect.Complex128, reflect.Complex64:
		r, err := strconv.ParseFloat(s[:strings.Index(s, ";")], 64)
		if err != nil {
			return -1, errors.New("Conversion error: " + err.Error())
		}
		im, err2 := strconv.ParseFloat(s[strings.Index(s, ";")+1:], 64)
		if err2 != nil {
			return -1, errors.New("Conversion error: " + err.Error())
		}
		val.SetComplex(complex(r, im))
	}
	return at, nil
}

func fieldValToRaw(val reflect.Value, typ reflect.Type) []byte {
	return append([]byte(valToString(val, typ)), byte(','))
}

func valToString(val reflect.Value, typ reflect.Type) string {
	switch typ.Kind() {
	case reflect.String:
		return val.String()
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int8:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Float32:
		return strconv.FormatFloat(val.Float(), byte('f'), -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(val.Float(), byte('f'), -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(val.Bool())
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		return strconv.FormatUint(val.Uint(), 10)
	case reflect.Complex128, reflect.Complex64:
		return strconv.FormatFloat(real(val.Complex()), byte('f'), -1, 64) + ";" +
			strconv.FormatFloat(imag(val.Complex()), byte('f'), -1, 64)
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
