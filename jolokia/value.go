package jolokia

import (
	"math"
	"reflect"
	"encoding/json"
	"errors"
	"strings"
	"github.com/iancoleman/strcase"
)

var (
	floatType = reflect.TypeOf(float64(0))
	stringType = reflect.TypeOf("")

	errNotAFloat = errors.New("value is not a float")
)

// toFloat converts a given interface to a float64 value.
// Code from https://stackoverflow.com/questions/20767724/converting-unknown-interface-to-float64-in-golang
// working example at https://play.golang.org/p/v-QrbeOWtz
func toFloat(unk interface{}) (float64, error) {
	switch i := unk.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case string:
		return math.NaN(), errNotAFloat
	default:
		v := reflect.ValueOf(unk)
		v = reflect.Indirect(v)
		if v.Type().ConvertibleTo(floatType) {
			fv := v.Convert(floatType)
			return fv.Float(), nil
		} else {
			return math.NaN(), errNotAFloat
		}
	}
}

func getValues(target string, msg json.RawMessage) (map[string]float64, error) {
	result := make(map[string]float64, 0)

	var value NestedValue
	if err := json.Unmarshal(msg, &value); err == nil {
		for key, val := range value {
			nestedResult, err := getValues(strings.Join([]string{target, strcase.ToSnake(key)}, "_"), val)
			if err != nil {
				return nil, err
			}

			result = mergeMaps(result, nestedResult)
		}
	}

	val, err := getFloatValue(msg)
	if err != nil {
		// if the value is not parseable as float and is not a nested value an empty map is returned
		if err == errNotAFloat {
			return result, nil
		}

		return nil, err
	}

	result[target] = val

	return result, nil
}

func getFloatValue(msg json.RawMessage) (float64, error) {
	var value SimpleValue
	if err := json.Unmarshal(msg, &value); err != nil {
		return 0, err
	}

	return toFloat(value)
}

func mergeMaps(v1, v2 map[string]float64) map[string]float64 {
	for k, v := range v2 {
		v1[k] = v
	}

	return v1
}
