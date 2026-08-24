package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrNotAStruct  = errors.New("value is not a struct")
	ErrInvalidLen  = errors.New("invalid length")
	ErrRegexpMatch = errors.New("does not match regular expression")
	ErrNotInSet    = errors.New("value is not in allowed set")
	ErrMinBound    = errors.New("value is below minimum bound")
	ErrMaxBound    = errors.New("value is above maximum bound")
)

type ValidationError struct {
	Field string
	Err   error
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("field %s: %v", e.Field, e.Err)
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for i, err := range v {
		if i > 0 {
			sb.WriteString("; ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

type ProgramError struct {
	Err error
}

func (p ProgramError) Error() string { return fmt.Sprintf("validator program error: %v", p.Err) }
func (p ProgramError) Unwrap() error { return p.Err }

//nolint:lll,nestif
func Validate(v interface{}) error {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return ErrNotAStruct
	}

	valueType := value.Type()
	var validationErrors ValidationErrors

	for i := 0; i < value.NumField(); i++ {
		fieldValue := value.Field(i)
		fieldType := valueType.Field(i)

		// Игнорируем неэкспортируемые (приватные) поля
		if !fieldValue.CanInterface() {
			continue
		}

		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}

		rules := strings.Split(tag, "|")

		// Обработка слайсов ([]int, []string)
		if fieldValue.Kind() == reflect.Slice {
			// поддержка слайсов любых поддерживаемых базовых типов (поддерживаем []int, []string +алиасы на них)
			elemKind := fieldValue.Type().Elem().Kind()
			if elemKind != reflect.String && elemKind != reflect.Int && elemKind != reflect.Int8 && elemKind != reflect.Int16 && elemKind != reflect.Int32 && elemKind != reflect.Int64 {
				return ProgramError{
					Err: fmt.Errorf("unsupported slice element type %s for field %s", elemKind, fieldType.Name),
				}
			}

			for j := 0; j < fieldValue.Len(); j++ {
				elem := fieldValue.Index(j)
				fieldNameWithIndex := fmt.Sprintf("%s[%d]", fieldType.Name, j)

				if err := validateValue(elem, rules, fieldNameWithIndex); err != nil {
					var programError ProgramError
					if errors.As(err, &programError) {
						return programError
					}
					var vErrors ValidationErrors
					if errors.As(err, &vErrors) {
						validationErrors = append(validationErrors, vErrors...)
					}
				}
			}
			continue
		}

		// Обработка одиночных значений (int, string)
		if err := validateValue(fieldValue, rules, fieldType.Name); err != nil {
			var programError ProgramError
			if errors.As(err, &programError) {
				return programError
			}
			var vErrors ValidationErrors
			if errors.As(err, &vErrors) {
				validationErrors = append(validationErrors, vErrors...)
			}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}

	return nil
}

//nolint:gocognit,lll
func validateValue(val reflect.Value, rules []string, fieldName string) error {
	var validationErrors ValidationErrors

	kind := val.Kind()

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)
		if rule == "" {
			continue
		}

		parts := strings.SplitN(rule, ":", 2)
		if len(parts) != 2 {
			return ProgramError{Err: fmt.Errorf("malformed rule %q", rule)}
		}

		ruleName := parts[0]
		ruleVal := parts[1]

		//nolint:exhaustive
		switch kind {
		case reflect.String:
			str := val.String()
			switch ruleName {
			case "len":
				expectedLen, err := strconv.Atoi(ruleVal)
				if err != nil {
					return ProgramError{Err: fmt.Errorf("invalid rule len value %q: %w", ruleVal, err)}
				}
				if len(str) != expectedLen {
					validationErrors = append(validationErrors, ValidationError{Field: fieldName, Err: fmt.Errorf("%w: expected %d, got %d", ErrInvalidLen, expectedLen, len(str))})
				}
			case "regexp":
				re, err := regexp.Compile(ruleVal)
				if err != nil {
					return ProgramError{Err: fmt.Errorf("invalid regexp %q: %w", ruleVal, err)}
				}
				if !re.MatchString(str) {
					validationErrors = append(validationErrors, ValidationError{Field: fieldName, Err: fmt.Errorf("%w: must match /%s/", ErrRegexpMatch, ruleVal)})
				}
			case "in":
				allowed := strings.Split(ruleVal, ",")
				found := false
				for _, a := range allowed {
					if str == a {
						found = true
						break
					}
				}
				if !found {
					validationErrors = append(validationErrors, ValidationError{Field: fieldName, Err: fmt.Errorf("%w: missing from [%s]", ErrNotInSet, ruleVal)})
				}
			default:
				return ProgramError{Err: fmt.Errorf("unsupported rule %q for string type", ruleName)}
			}

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			num := val.Int()
			switch ruleName {
			case "min":
				minVal, err := strconv.ParseInt(ruleVal, 10, 64)
				if err != nil {
					return ProgramError{Err: fmt.Errorf("invalid rule min value %q: %w", ruleVal, err)}
				}
				if num < minVal {
					validationErrors = append(validationErrors, ValidationError{Field: fieldName, Err: fmt.Errorf("%w: minimum is %d, got %d", ErrMinBound, minVal, num)})
				}
			case "max":
				maxVal, err := strconv.ParseInt(ruleVal, 10, 64)
				if err != nil {
					return ProgramError{Err: fmt.Errorf("invalid max value %q: %w", ruleVal, err)}
				}
				if num > maxVal {
					validationErrors = append(validationErrors, ValidationError{Field: fieldName, Err: fmt.Errorf("%w: maximum is %d, got %d", ErrMaxBound, maxVal, num)})
				}
			case "in":
				allowedParts := strings.Split(ruleVal, ",")
				found := false
				for _, p := range allowedParts {
					allowedNum, err := strconv.ParseInt(p, 10, 64)
					if err != nil {
						return ProgramError{Err: fmt.Errorf("invalid number %q in set: %w", p, err)}
					}
					if num == allowedNum {
						found = true
						break
					}
				}
				if !found {
					validationErrors = append(validationErrors, ValidationError{Field: fieldName, Err: fmt.Errorf("%w: missing from [%s]", ErrNotInSet, ruleVal)})
				}
			default:
				return ProgramError{Err: fmt.Errorf("unsupported rule %q for int", ruleName)}
			}

		default:
			return ProgramError{Err: fmt.Errorf("unsupported type %s for field %s", kind, fieldName)}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}
