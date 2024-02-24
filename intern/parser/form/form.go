package form

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

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

		if fieldName != "" {
			value, exists := input[fieldName]
			if exists && len(value) > 0 {
				// NOTE: Take only the first value.
				fieldValRaw := value[0]
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
				}
			}
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
