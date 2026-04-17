package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func FormatTime(t time.Time, format string) string {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return t.Format(format)
}

func CurrentTime(format string) string {
	return FormatTime(time.Now(), format)
}

func ValidateConfig(config map[string]interface{}, requiredKeys []string) error {
	missingKeys := make([]string, 0)
	for _, key := range requiredKeys {
		if _, exists := config[key]; !exists {
			missingKeys = append(missingKeys, key)
		}
	}
	if len(missingKeys) > 0 {
		return fmt.Errorf("config missing required keys: %v", missingKeys)
	}
	return nil
}

func SafeImport(modulePath, funcName string) (interface{}, error) {
	parts := strings.Split(modulePath, "/")
	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid module path")
	}

	moduleName := parts[len(parts)-1]

	switch moduleName {
	case "json":
		if funcName == "" || funcName == "Marshal" {
			return jsonMarshal, nil
		} else if funcName == "Unmarshal" {
			return jsonUnmarshal, nil
		}
	case "os":
		if funcName == "Getenv" {
			return osGetenv, nil
		}
	}

	return nil, fmt.Errorf("module %s or function %s not found", moduleName, funcName)
}

func jsonMarshal(v interface{}) (interface{}, error) {
	return v, nil
}

func jsonUnmarshal(data string, v interface{}) error {
	return fmt.Errorf("not implemented")
}

func osGetenv(key string) string {
	return os.Getenv(key)
}

func EnsureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func GetProjectRoot() string {
	execPath, _ := os.Executable()
	dir := filepath.Dir(execPath)
	return dir
}

func GetCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func WriteFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

func GetEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		var result float64
		if _, err := fmt.Sscanf(value, "%f", &result); err == nil {
			return result
		}
	}
	return defaultValue
}

type Result struct {
	Value interface{}
	Error error
}

func Ok(val interface{}) Result {
	return Result{Value: val, Error: nil}
}

func Err(err error) Result {
	return Result{Value: nil, Error: err}
}

func (r Result) IsOk() bool {
	return r.Error == nil
}

func (r Result) Or(defaultVal interface{}) interface{} {
	if r.Error != nil {
		return defaultVal
	}
	return r.Value
}
