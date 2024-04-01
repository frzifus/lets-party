package form

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// Unmarshal parses the url.Values data and stores the result
// in the value pointed to by target. If target is nil or not a pointer,
// Unmarshal returns an [InvalidUnmarshalError].
//
// Unmarshal is currently still limited. Given list values will be ignored and
// the only supported slice type is string.
func Unmarshal(input url.Values, target any) error {
	val := reflect.ValueOf(target)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return &InvalidUnmarshalError{Type: reflect.TypeOf(target)}
	}

	v := val.Elem()
	ttype := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := ttype.Field(i)
		fieldName := field.Tag.Get("form")

		value, _ := input[fieldName]
		if fieldName == "" || fieldName == "-" {
			continue
		}
		var fieldValRaw string
		if len(value) > 0 {
			// TODO: Respect all values!
			fieldValRaw = value[0]
		}
		fieldVal := v.Field(i)
		// TODO: support all types.
		switch field.Type.Kind() {
		case reflect.String:
			fieldVal.SetString(fieldValRaw)
		case reflect.Bool:
			boolValue := strings.ToLower(fieldValRaw) == "true"
			fieldVal.SetBool(boolValue)
		case reflect.Int:
			if fieldValRaw == "" {
				continue
			}
			intValue, err := strconv.Atoi(fieldValRaw)
			if err != nil {
				return err
			}
			fieldVal.SetInt(int64(intValue))
		case reflect.Slice:
			// NOTE: We assume the slice is of type string.
			// if not... we panic... because we can.
			// TODO: Create reflect.Slice and call Unmarshal.
			fieldVal.Set(reflect.AppendSlice(fieldVal, reflect.ValueOf(value)))
		case reflect.Struct:
			newInput := make(url.Values, len(input))
			for k, v := range input {
				newInput[strings.TrimPrefix(k, fmt.Sprintf("%s.", fieldName))] = v
			}

			if err := Unmarshal(newInput, fieldVal.Addr().Interface()); err != nil {
				return err
			}
		default:
			panic(fmt.Sprintf("unsupported type: %s", field.Type.String()))
		}
	}
	return nil
}

type InvalidUnmarshalError struct {
	Type reflect.Type
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "form: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Pointer {
		return "form: Unmarshal(non-pointer " + e.Type.String() + ")"
	}
	return "form: Unmarshal(nil " + e.Type.String() + ")"
}
