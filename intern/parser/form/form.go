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
// Unmarshal is currently still to a few primitive types and basic recursion.
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

		value := input[fieldName]
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
		case reflect.Float64:
			if fieldValRaw == "" {
				continue
			}
			fValue, err := strconv.ParseFloat(fieldValRaw, 64)
			if err != nil {
				return err
			}
			fieldVal.SetFloat(fValue)
		case reflect.Slice:
			sliceValue := reflect.ValueOf(value)
			for i := 0; i < sliceValue.Len(); i++ {
				if isPrimitiveType(sliceValue.Type().Elem().Kind()) {
					fieldVal.Set(reflect.Append(fieldVal, sliceValue.Index(i)))
					continue
				}

				var ptr any
				if sliceValue.Kind() == reflect.Ptr {
					ptr = sliceValue.Index(i).Interface()
				} else {
					ptr = sliceValue.Index(i).Addr().Interface()
				}
				if err := Unmarshal(input, ptr); err != nil {
					return err
				}
			}
		case reflect.Struct:
			newInput := make(url.Values, len(input))
			for k, v := range input {
				prefix := fmt.Sprintf("%s.", fieldName)
				if strings.HasPrefix(k, prefix) {
					newInput[strings.TrimPrefix(k, prefix)] = v
				}
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

func isPrimitiveType(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	default:
		return false
	}
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
