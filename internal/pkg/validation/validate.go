package validation

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"bitbucket.org/Amartha/go-accounting/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-multierror"
)

var validate = validator.New()

func init() {
	registerNoSpecialCharacters()
	registerNoSpacesAtStartOrEnd()
	registerDate()
	registerDatetime()
	registerAlphanumericMix()
	registerAlphanumDashUscore()
	registerMaxWithoutSpaces()
	registerMinWithoutSpaces()
}

func ValidateStruct(toValidate interface{}) error {
	// register function to get tag name from json tags.
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	var errs *multierror.Error
	if err := validate.Struct(toValidate); err != nil {
		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		if _, ok := err.(*validator.InvalidValidationError); ok {
			errs = multierror.Append(errs, ErrorValidateResponse{
				Message: err.Error(),
			})
			return errs.ErrorOrNil()
		}

		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			for _, valErr := range valErrs {
				key := fmt.Sprintf("%s_%s", valErr.Namespace(), valErr.Tag())
				data, found := models.MapErrors[key]
				if found {
					errorResponse := ErrorValidateResponse{
						Code:    data.Code,
						Field:   valErr.Field(),
						Message: data.ErrorMessage.Error(),
					}
					errs = multierror.Append(errs, errorResponse)
				} else {
					key := fmt.Sprintf("%s_%s", valErr.Field(), valErr.Tag())
					if data, found := models.MapErrors[key]; found {
						errorResponse := ErrorValidateResponse{
							Code:    data.Code,
							Field:   valErr.Field(),
							Message: data.ErrorMessage.Error(),
						}
						errs = multierror.Append(errs, errorResponse)
					} else {
						errorResponse := ErrorValidateResponse{
							Code:    "UNKNOW",
							Field:   valErr.Field(),
							Message: strings.TrimSpace(fmt.Sprintf("%s %s", valErr.Tag(), valErr.Param())),
						}
						errs = multierror.Append(errs, errorResponse)
					}
				}
			}
		}
	}

	return errs.ErrorOrNil()
}

func registerNoSpecialCharacters() {
	validate.RegisterValidation("nospecial", func(fl validator.FieldLevel) bool {
		input := fl.Field().String()
		// Define a regular expression pattern that allows only letters and digits.
		// Allow space
		pattern := "^[a-zA-Z0-9 ]*$"
		return regexp.MustCompile(pattern).MatchString(input)
	})
}

func registerNoSpacesAtStartOrEnd() {
	validate.RegisterValidation("noStartEndSpaces", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		return str == "" || (str[0] != ' ' && str[len(str)-1] != ' ')
	})
}

func registerDate() {
	validate.RegisterValidation("date", func(fl validator.FieldLevel) bool {
		input := fl.Field().String()
		pattern := `\d{4}-\d{2}-\d{2}`
		return regexp.MustCompile(pattern).MatchString(input)
	})
}

func registerDatetime() {
	validate.RegisterValidation("datetime", func(fl validator.FieldLevel) bool {
		input := fl.Field().String()
		pattern := `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`
		return regexp.MustCompile(pattern).MatchString(input)
	})
}

func registerAlphanumericMix() {
	validate.RegisterValidation("alphanumericMix", func(fl validator.FieldLevel) bool {
		input := fl.Field().String()
		// Define a regular expression patterns that only allow these characters:
		// letters (a-z, A-Z), digits (0-9), spaces ( ), parentheses [()], dash/hyphen/minus (-),
		// slash (/), dot (.), comma (,), colon (:), semicolon (;), and quotation marks (',",“”,‘’)
		pattern := `^[a-zA-Z0-9 \-()\/.,:;'"“”‘’]*$`
		return regexp.MustCompile(pattern).MatchString(input)
	})
}

func registerAlphanumDashUscore() {
	validate.RegisterValidation("alphanumDashUscore", func(fl validator.FieldLevel) bool {
		input := fl.Field().String()
		// Define a regular expression patterns that only allow these characters:
		// letters (a-z, A-Z), digits (0-9), dash/hyphen/minus (-), and underscore (_)
		pattern := `^[a-zA-Z0-9\-_]*$`
		return regexp.MustCompile(pattern).MatchString(input)
	})
}

func registerMaxWithoutSpaces() {
	validate.RegisterValidation("maxnospace", func(fl validator.FieldLevel) bool {
		param := fl.Param()
		desiredLen, err := strconv.Atoi(param)
		if err != nil {
			return false
		}

		str := fl.Field().String()
		noSpaces := strings.ReplaceAll(str, " ", "")
		return len(noSpaces) <= desiredLen
	})
}

func registerMinWithoutSpaces() {
	validate.RegisterValidation("minnospace", func(fl validator.FieldLevel) bool {
		param := fl.Param()
		desiredLen, err := strconv.Atoi(param)
		if err != nil {
			return false
		}

		str := fl.Field().String()
		noSpaces := strings.ReplaceAll(str, " ", "")
		return len(noSpaces) >= desiredLen
	})
}
