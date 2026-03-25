package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"sync"
	"time"

	"github.com/clineomx/trussrod/apperr"
	"github.com/go-playground/validator/v10"
)

func isDate(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

func minDate(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	minDateStr := fl.Param()

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return false
	}

	m, err := time.Parse("2006-01-02", minDateStr)
	if err != nil {
		return false
	}

	return date.After(m) || date.Equal(m)
}

func isPhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	clean := regexp.MustCompile(`[\s\-\(\)]`).ReplaceAllString(phone, "")

	p1 := `^\+52\d{10}$`
	p2 := `^\d{10}$`
	m1, _ := regexp.MatchString(p1, clean)
	m2, _ := regexp.MatchString(p2, clean)
	return m1 || m2
}

func isSSN(fl validator.FieldLevel) bool {
	nss := fl.Field().String()
	matched, _ := regexp.MatchString(`^\d{11}$`, nss)
	return matched
}

func isTaxID(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	m, _ := regexp.MatchString(`^[A-ZÑ&]{4}\d{6}[A-Z0-9]{3}$`, value)
	return m
}

func isCitizenID(fl validator.FieldLevel) bool {
	value := fl.Field().String()

	if len(value) != 18 {
		return false
	}

	matched, _ := regexp.MatchString(`^[A-Z]{4}\d{6}[HM][A-Z]{5}[0-9A-Z]\d$`, value)
	return matched
}

var (
	validate     *validator.Validate
	validateOnce sync.Once
)

func get() *validator.Validate {
	validateOnce.Do(func() {
		validate = validator.New()

		validate.RegisterValidation("validdate", isDate)
		validate.RegisterValidation("mindate", minDate)
		validate.RegisterValidation("phone", isPhone)
		validate.RegisterValidation("ssn", isSSN)
		validate.RegisterValidation("tax_id", isTaxID)
		validate.RegisterValidation("citizen_id", isCitizenID)
	})

	return validate
}

func ValidatePayload(obj any) error {
	val := reflect.ValueOf(obj)

	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return apperr.BadRequest("cannot validate nil pointer")
		}
		val = val.Elem()
	}

	// Intentar validar
	toValidate := val.Interface()

	err := get().Struct(toValidate)

	if err != nil {
		validationErrors := err.(validator.ValidationErrors)
		fieldErrors := make(map[string]string)
		var details []string

		for _, ve := range validationErrors {
			fieldName := ve.Field()
			tag := ve.Tag()

			// Create a more descriptive error message
			var errorMsg string
			switch tag {
			case "required":
				errorMsg = fmt.Sprintf("The field '%s' is required", fieldName)
			case "email":
				errorMsg = fmt.Sprintf("The field '%s' must be a valid email address", fieldName)
			case "min":
				errorMsg = fmt.Sprintf("The field '%s' must be at least %s characters long", fieldName, ve.Param())
			case "max":
				errorMsg = fmt.Sprintf("The field '%s' must be at most %s characters long", fieldName, ve.Param())
			case "validdate":
				errorMsg = fmt.Sprintf("The field '%s' must be a valid date in YYYY-MM-DD format", fieldName)
			case "mindate":
				errorMsg = fmt.Sprintf("The field '%s' must be on or after %s", fieldName, ve.Param())
			case "phone":
				errorMsg = fmt.Sprintf("The field '%s' must be a valid phone number", fieldName)
			case "ssn":
				errorMsg = fmt.Sprintf("The field '%s' must be a valid SSN (11 digits)", fieldName)
			case "tax_id":
				errorMsg = fmt.Sprintf("The field '%s' must be a valid tax ID (RFC format)", fieldName)
			case "citizen_id":
				errorMsg = fmt.Sprintf("The field '%s' must be a valid citizen ID (CURP format)", fieldName)
			default:
				errorMsg = fmt.Sprintf("The field '%s' failed validation for tag '%s'", fieldName, tag)
			}

			fieldErrors[fieldName] = errorMsg
			details = append(details, errorMsg)
		}

		// Use structured field errors for better API responses
		return apperr.BadRequest("invalid payload")
	}
	return nil
}
