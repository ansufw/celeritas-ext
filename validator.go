package celeritas

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
)

type Validator struct {
	Data   url.Values
	Errors map[string]string
}

func (c *Celeritas) Validator(data url.Values) *Validator {
	return &Validator{
		Data:   data,
		Errors: make(map[string]string),
	}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

func (v *Validator) Has(field string, r *http.Request) bool {
	x := r.Form.Get(field)
	if x == "" {
		return false
	}
	return true
}

func (v *Validator) Required(r *http.Request, fields ...string) {
	for _, field := range fields {
		value := r.Form.Get(field)
		if strings.TrimSpace(value) == "" {
			v.AddError(field, "This field cannot be blank")
		}
	}
}

func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

func (v *Validator) IsEmail(field, value string) {
	if !govalidator.IsEmail(value) {
		v.AddError(field, "This field must be a valid email address")
	}
}

func (v *Validator) IsInt(field, value string) {
	_, err := strconv.Atoi(value)
	if err != nil {
		v.AddError(field, "This field must be an integer")
	}
}

func (v *Validator) IsFloat(field, value string) {
	_, err := strconv.ParseFloat(value, 64)
	if err != nil {
		v.AddError(field, "This field must be a float")
	}
}

func (v *Validator) IsDateISO(field, value string) {
	_, err := time.Parse("2006-01-02", value)
	if err != nil {
		v.AddError(field, "This field must be a valid date in the form of YYYY-MM-DD")
	}
}

func (v *Validator) NoSpaces(field, value string) {
	if govalidator.HasWhitespace(value) {
		v.AddError(field, "This field cannot contain spaces")
	}
}

func (v *Validator) MaxChars(r *http.Request, field string, max int) {
	value := r.Form.Get(field)
	if strings.TrimSpace(value) == "" {
		return
	}
	if len(value) > max {
		v.AddError(field, fmt.Sprintf("This field cannot be longer than %d characters", max))
	}
}
