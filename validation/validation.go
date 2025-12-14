package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Domedik/trussrod/errors"
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
			return errors.BadRequest("cannot validate nil pointer")
		}
		val = val.Elem()
	}

	// Intentar validar
	toValidate := val.Interface()

	err := get().Struct(toValidate)
	var v []string

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			msg := fmt.Sprintf("Field %s failed on the '%s' tag", err.Field(), err.Tag())
			v = append(v, msg)
		}
		return errors.ValidationFailed(strings.Join(v, ","))
	}
	return nil
}
