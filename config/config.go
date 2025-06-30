package config

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"

	"github.com/spf13/viper"
)

const (
	defaultTagName = "default"
)

// ReadConfig 读取配置文件，支持'$'及'${}'环境变量
// 支持通过default tag定义默认值
//
//	type ConfigTest2 struct {
//		Int           int
//		String        string
//		IntDefault    int    `default:"1"`
//		StringDefault string `default:"string_default"`
//	}
//
// 如果 IntDefault，StringDefault在配置文件中没有提供，则会使用1和string_default
func ReadConfig(file string, config interface{}) (err error) {
	ext := getConfigType(path.Ext(file))
	viper.SetConfigType(ext)
	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	// 替换环境变量
	configRaw := []byte(os.ExpandEnv(string(content)))
	if err = viper.ReadConfig(bytes.NewBuffer(configRaw)); err != nil {
		return
	}
	if err := viper.Unmarshal(config); err != nil {
		return err
	}
	return parseDefaultTag(config)
}

func getConfigType(ext string) string {
	switch ext {
	case ".yml", ".yaml":
		return "yaml"
	case ".json":
		return "json"
	case ".ini":
		return "ini"
	case ".toml":
		return "toml"
	default:
		panic("unknown file type")
	}
}

func parseDefaultTag(config interface{}) error {
	reflectValue := reflect.ValueOf(config)
	if reflectValue.Kind() != reflect.Ptr || reflectValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expected a pointer to a struct")
	}
	return parseTag(reflectValue.Elem(), defaultTagName)
}

func parseTag(val reflect.Value, tagName string) (err error) {
	typ := val.Type()
	switch typ.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			fieldValue := val.Field(i)
			fieldType := typ.Field(i)
			if !fieldValue.CanSet() {
				continue
			}
			switch fieldValue.Kind() {
			case reflect.Struct, reflect.Array, reflect.Slice, reflect.Map, reflect.Interface, reflect.Ptr:
				if err := parseTag(fieldValue, tagName); err != nil {
					return err
				}
			default:
				tagValue := fieldType.Tag.Get(tagName)
				if tagValue != "" && isZero(fieldValue) {
					if err := setFieldValue(fieldValue, tagValue); err != nil {
						return fmt.Errorf("error setting value for field %s using tag '%s': %w", fieldType.Name, tagName, err)
					}
				}
			}
		}
	case reflect.Ptr, reflect.Interface:
		if !val.IsNil() {
			if err := parseTag(val.Elem(), tagName); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for j := 0; j < val.Len(); j++ {
			elem := val.Index(j)
			if err := parseTag(elem, tagName); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			elem := val.MapIndex(key)
			if err := parseTag(elem, tagName); err != nil {
				return err
			}
		}
	default:
	}
	return
}

func isZero(v reflect.Value) bool {
	return v.Interface() == reflect.Zero(v.Type()).Interface()
}

func setFieldValue(field reflect.Value, tag string) (err error) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(tag)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(tag, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(tag, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(tag, 64)
		if err != nil {
			return err
		}
		field.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(tag)
		if err != nil {
			return err
		}
		field.SetBool(v)
	default:
		return fmt.Errorf("unsupported type %s", field.Kind())
	}
	return err
}
