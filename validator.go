package goify

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type Validator struct {
	validators map[string]ValidatorFunc
}

type ValidatorFunc func(value interface{}, param string) error

type ValidationError struct {
	Field   string `json:"field"`
	Value   interface{} `json:"value"`
	Tag     string `json:"tag"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

func NewValidator() *Validator {
	v := &Validator{
		validators: make(map[string]ValidatorFunc),
	}

	v.registerBuiltinValidators()
	
	return v
}

var defaultValidator = NewValidator()

func (v *Validator) RegisterValidator(tag string, fn ValidatorFunc) {
	v.validators[tag] = fn
}

func RegisterValidator(tag string, fn ValidatorFunc) {
	defaultValidator.RegisterValidator(tag, fn)
}

func (v *Validator) Validate(s interface{}) ValidationErrors {
	return v.validateStruct(reflect.ValueOf(s), "")
}

func Validate(s interface{}) ValidationErrors {
	return defaultValidator.Validate(s)
}

func (v *Validator) validateStruct(val reflect.Value, prefix string) ValidationErrors {
	var errors ValidationErrors

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return errors
		}
		val = val.Elem()
	}
	
	if val.Kind() != reflect.Struct {
		return errors
	}
	
	typ := val.Type()
	
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanInterface() {
			continue
		}
		
		fieldName := fieldType.Name
		if prefix != "" {
			fieldName = prefix + "." + fieldName
		}

		if jsonTag := fieldType.Tag.Get("json"); jsonTag != "" {
			if tagName := strings.Split(jsonTag, ",")[0]; tagName != "" && tagName != "-" {
				if prefix != "" {
					fieldName = prefix + "." + tagName
				} else {
					fieldName = tagName
				}
			}
		}

		if field.Kind() == reflect.Struct || (field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct) {
			errors = append(errors, v.validateStruct(field, fieldName)...)
			continue
		}

		if field.Kind() == reflect.Slice {
			for j := 0; j < field.Len(); j++ {
				item := field.Index(j)
				if item.Kind() == reflect.Struct || (item.Kind() == reflect.Ptr && item.Type().Elem().Kind() == reflect.Struct) {
					indexFieldName := fmt.Sprintf("%s[%d]", fieldName, j)
					errors = append(errors, v.validateStruct(item, indexFieldName)...)
				}
			}
		}

		validateTag := fieldType.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		rules := strings.Split(validateTag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == "" {
				continue
			}

			parts := strings.SplitN(rule, "=", 2)
			tag := parts[0]
			param := ""
			if len(parts) > 1 {
				param = parts[1]
			}

			if validator, exists := v.validators[tag]; exists {
				if err := validator(field.Interface(), param); err != nil {
					errors = append(errors, ValidationError{
						Field:   fieldName,
						Value:   field.Interface(),
						Tag:     tag,
						Param:   param,
						Message: err.Error(),
					})
				}
			}
		}
	}
	
	return errors
}

func (v *Validator) registerBuiltinValidators() {
	v.RegisterValidator("required", func(value interface{}, param string) error {
		if isEmpty(value) {
			return fmt.Errorf("field is required")
		}
		return nil
	})

	v.RegisterValidator("min", func(value interface{}, param string) error {
		min, err := strconv.Atoi(param)
		if err != nil {
			return fmt.Errorf("invalid min parameter: %s", param)
		}
		
		switch v := value.(type) {
		case string:
			if len(v) < min {
				return fmt.Errorf("minimum length is %d", min)
			}
		case int, int8, int16, int32, int64:
			val := reflect.ValueOf(v).Int()
			if val < int64(min) {
				return fmt.Errorf("minimum value is %d", min)
			}
		case uint, uint8, uint16, uint32, uint64:
			val := reflect.ValueOf(v).Uint()
			if val < uint64(min) {
				return fmt.Errorf("minimum value is %d", min)
			}
		case float32, float64:
			val := reflect.ValueOf(v).Float()
			if val < float64(min) {
				return fmt.Errorf("minimum value is %d", min)
			}
		default:
			rv := reflect.ValueOf(v)
			if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
				if rv.Len() < min {
					return fmt.Errorf("minimum length is %d", min)
				}
			}
		}
		return nil
	})

	v.RegisterValidator("max", func(value interface{}, param string) error {
		max, err := strconv.Atoi(param)
		if err != nil {
			return fmt.Errorf("invalid max parameter: %s", param)
		}
		
		switch v := value.(type) {
		case string:
			if len(v) > max {
				return fmt.Errorf("maximum length is %d", max)
			}
		case int, int8, int16, int32, int64:
			val := reflect.ValueOf(v).Int()
			if val > int64(max) {
				return fmt.Errorf("maximum value is %d", max)
			}
		case uint, uint8, uint16, uint32, uint64:
			val := reflect.ValueOf(v).Uint()
			if val > uint64(max) {
				return fmt.Errorf("maximum value is %d", max)
			}
		case float32, float64:
			val := reflect.ValueOf(v).Float()
			if val > float64(max) {
				return fmt.Errorf("maximum value is %d", max)
			}
		default:
			rv := reflect.ValueOf(v)
			if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
				if rv.Len() > max {
					return fmt.Errorf("maximum length is %d", max)
				}
			}
		}
		return nil
	})

	v.RegisterValidator("email", func(value interface{}, param string) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("email validation only works on strings")
		}
		
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(str) {
			return fmt.Errorf("invalid email format")
		}
		return nil
	})

	v.RegisterValidator("url", func(value interface{}, param string) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("url validation only works on strings")
		}
		
		urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
		if !urlRegex.MatchString(str) {
			return fmt.Errorf("invalid URL format")
		}
		return nil
	})

	v.RegisterValidator("alpha", func(value interface{}, param string) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("alpha validation only works on strings")
		}
		
		alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
		if !alphaRegex.MatchString(str) {
			return fmt.Errorf("field must contain only letters")
		}
		return nil
	})
	
	// Alphanumeric validator
	v.RegisterValidator("alphanum", func(value interface{}, param string) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("alphanum validation only works on strings")
		}
		
		alphanumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
		if !alphanumRegex.MatchString(str) {
			return fmt.Errorf("field must contain only letters and numbers")
		}
		return nil
	})

	v.RegisterValidator("numeric", func(value interface{}, param string) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("numeric validation only works on strings")
		}
		
		numericRegex := regexp.MustCompile(`^[0-9]+$`)
		if !numericRegex.MatchString(str) {
			return fmt.Errorf("field must contain only numbers")
		}
		return nil
	})

	v.RegisterValidator("oneof", func(value interface{}, param string) error {
		str := fmt.Sprintf("%v", value)
		options := strings.Split(param, " ")
		
		for _, option := range options {
			if str == option {
				return nil
			}
		}
		
		return fmt.Errorf("field must be one of: %s", param)
	})
}

func isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		return rv.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}