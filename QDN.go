package qdn

import (
	"bytes"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

// Format puts the raw byte data in a more readable state for use in a text editor.
// Keep in mind that this uses system resources, so its not adviseable to use for network transmission
func Format(r []byte) ([]byte, error) {
	var formatted []byte
	var tabC int
	const tab byte = byte('\t')
	const nl byte = byte('\n')

	for i := 0; i < len(r); i++ {
		if r[i] == nl || r[i] == tab {
			return r, errors.New("Bytes were already formatted")
		}

		if r[i] == byte('<') {
			formatted = append(formatted, nl)
			formatted = append(formatted, createTabS(tabC)...)
			formatted = append(formatted, r[i], nl)
			tabC++
			formatted = append(formatted, createTabS(tabC)...)
			continue
		}
		if r[i] == byte('>') {
			formatted = append(formatted[:len(formatted)-1], r[i])
			tabC--
			continue
		}

		if r[i] == byte(',') {
			formatted = append(formatted, r[i], nl)
			formatted = append(formatted, createTabS(tabC)...)
			continue
		}

		formatted = append(formatted, r[i])
	}
	return formatted, nil
}

func createTabS(tabC int) []byte {
	var tabS []byte
	for j := 0; j < tabC; j++ {
		tabS = append(tabS, byte('\t'))
	}
	return tabS
}

// Unmarshal fills a given interface with the corresponding qdn byte data.
func Unmarshal(stru interface{}, data []byte) error {
	if err := unmarshalInitErrors(stru, data); err != nil {
		return err
	}

	count := reflect.TypeOf(stru).Elem().NumField()
	if count < 1 {
		return errors.New("qdn.Unmarshal error: The struct does not contain any fields")
	}

	s := reflect.ValueOf(stru).Elem()

	var (
		at  int
		err error
	)
	for i := 0; i < count; i++ {
		if reflect.TypeOf(stru).Elem().Field(i).Type.Kind() == reflect.Struct {
			c := bytes.Count(data[at+1:], []byte{byte('<')})
			err = Unmarshal(s.Field(i).Addr().Interface(), data[at+1:2+at+allIndizes(data[at+1:], byte('>'))[c-1]])
			at += allIndizes(data[at+1:], byte('>'))[c-1] + 2
			if err != nil {
				return err
			}
			continue
		}

		at, err = strToVal(s.Field(i), reflect.TypeOf(stru).Elem().Field(i).Type, data, at)
		if err != nil {
			return err
		}
	}

	return nil
}

// Marshal turns the given struct into a raw byte array
func Marshal(stru interface{}) ([]byte, error) {
	if reflect.TypeOf(stru).Kind() != reflect.Struct {
		return []byte{}, errors.New("qdn.Marshal error: Provided parameter is no struct")
	}

	count := reflect.TypeOf(stru).NumField()
	if count < 1 {
		return []byte{}, errors.New("qdn.Marshal error: The struct does not contain any fields")
	}

	r := make([]byte, 0)
	r = append(r, setupRaw(reflect.TypeOf(stru))...)
	for i := 0; i < count; i++ {
		if reflect.TypeOf(stru).Field(i).Type.Kind() == reflect.Struct {
			raw, err := Marshal(reflect.ValueOf(stru).Field(i).Interface())
			if err != nil {
				return raw, err
			}
			r = append(r, raw...)
			r = append(r, byte(','))
			continue
		}
		r = append(r, fieldToRaw(reflect.ValueOf(stru).Field(i), reflect.TypeOf(stru).Field(i))...)
	}

	return append(r, byte('>')), nil
}

// allIndizes gets the indizes i of all occasions of a byte in a byte slice
func allIndizes(d []byte, b byte) []int {
	var ins []int
	for i, v := range d {
		if v == b {
			ins = append(ins, i)
		}
	}
	return ins
}

// strToVal turns a string(of a raw slice) into the values for the struct
func strToVal(val reflect.Value, typ reflect.Type, data []byte, at int) (int, error) {
	i := at + bytes.IndexRune(data[at:], '=')
	at = i + bytes.IndexRune(data[i:], ',')
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

// fieldToRaw turns the field name and the value into a raw slice
func fieldToRaw(val reflect.Value, typ reflect.StructField) []byte {
	r := append([]byte(typ.Name), byte('='))
	r = append(r, []byte(valToString(val, typ.Type))...)
	return append(r, byte(','))
}

// valToString turns the actual value to a string
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

// checks for errors at the beginning of a Unmarshal function
func unmarshalInitErrors(stru interface{}, data []byte) error {
	if reflect.TypeOf(stru).Kind() != reflect.Ptr {
		return errors.New("qdn.Unmarshal error: Provided parameter is no pointer")
	}

	if reflect.ValueOf(stru).IsNil() {
		return errors.New("qdn.Unmarshal error: Pointer is nil")
	}

	if !bytes.ContainsAny(data, reflect.TypeOf(stru).Elem().Name()) {
		return errors.New("qdn.Unmarshal error: Missmatch: data does not represent the struct")
	}
	return nil
}

// initially sets up a struct raw slice
func setupRaw(struName reflect.Type) []byte {
	return append([]byte(struName.Name()), byte('<'))
}
