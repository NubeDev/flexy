package common

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	"github.com/go-playground/locales/zh_Hant"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"strings"
)

var (
	Uni      *ut.UniversalTranslator
	Validate *validator.Validate
)

func InitValidate() {
	e := en.New()
	z := zh.New()
	zt := zh_Hant.New()
	Uni = ut.New(e, z, zt)
	Validate = validator.New()
}

func CheckBindStructParameter(s interface{}, c *gin.Context) (error, string) {
	v, _ := c.Get("trans")

	trans, ok := v.(ut.Translator)
	if !ok {
		trans, _ = Uni.GetTranslator("zh")
	}

	err := Validate.Struct(s)
	if err != nil {
		var errs validator.ValidationErrors
		errors.As(err, &errs)
		var sliceErrs []string
		for _, e := range errs {
			sliceErrs = append(sliceErrs, e.Translate(trans))
		}
		return errs, strings.Join(sliceErrs, ",")
	}

	return nil, ""
}
