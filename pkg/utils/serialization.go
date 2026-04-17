package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

func SerializeObject(obj interface{}, format string) (string, error) {
	switch format {
	case "json":
		data, err := json.MarshalIndent(obj, "", "  ")
		if err != nil {
			return "", fmt.Errorf("JSON marshal error: %w", err)
		}
		return string(data), nil
	case "compact":
		data, err := json.Marshal(obj)
		if err != nil {
			return "", fmt.Errorf("JSON marshal error: %w", err)
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("unsupported serialization format: %s", format)
	}
}

func DeserializeObject(data string, format string, target interface{}) error {
	switch format {
	case "json":
		if err := json.Unmarshal([]byte(data), target); err != nil {
			return fmt.Errorf("JSON unmarshal error: %w", err)
		}
		return nil
	default:
		return fmt.Errorf("unsupported deserialization format: %s", format)
	}
}

func SaveToFile(obj interface{}, filepath, format string) error {
	data, err := SerializeObject(obj, format)
	if err != nil {
		return err
	}

	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if format == "json" {
		flag |= os.O_TRUNC
	}

	file, err := os.OpenFile(filepath, flag, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(data)
	return err
}

func LoadFromFile(filepath, format string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	return result, nil
}

func ToJSON(obj interface{}) string {
	data, _ := json.MarshalIndent(obj, "", "  ")
	return string(data)
}

func FromJSON(data string, target interface{}) error {
	return json.Unmarshal([]byte(data), target)
}

func DeepCopy(src interface{}) (dst interface{}) {
	data, _ := json.Marshal(src)
	json.Unmarshal(data, &dst)
	return
}

func MergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range map1 {
		result[k] = v
	}

	for k, v := range map2 {
		if v1, ok := result[k].(map[string]interface{}); ok {
			if v2, ok := v.(map[string]interface{}); ok {
				result[k] = MergeMaps(v1, v2)
			} else {
				result[k] = v
			}
		} else {
			result[k] = v
		}
	}

	return result
}

func GetField(obj interface{}, fieldName string) (interface{}, bool) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, false
	}

	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil, false
	}

	return field.Interface(), true
}
