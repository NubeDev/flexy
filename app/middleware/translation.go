package middleware

import (
	"github.com/NubeDev/flexy/common"
	"github.com/gin-gonic/gin"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	zh_tw_translations "github.com/go-playground/validator/v10/translations/zh_tw"
)

func TranslationHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := c.DefaultQuery("locale", "en")
		trans, _ := common.Uni.GetTranslator(locale)
		switch locale {
		case "zh":
			zh_translations.RegisterDefaultTranslations(common.Validate, trans)
			break
		case "en":
			en_translations.RegisterDefaultTranslations(common.Validate, trans)
			break
		case "zh_tw":
			zh_tw_translations.RegisterDefaultTranslations(common.Validate, trans)
			break
		default:
			zh_translations.RegisterDefaultTranslations(common.Validate, trans)
			break
		}

		c.Set("trans", trans)
		c.Next()
	}
}
