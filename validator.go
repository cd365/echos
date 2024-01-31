package echos

import (
	"reflect"
	"regexp"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	Validator *validator.Validate
}

func (s *Validator) Validate(i interface{}) error {
	return s.Validator.Struct(i)
}

func NewValidator() (validate *Validator, err error) {
	validate = &Validator{
		Validator: validator.New(),
	}

	/*
	 * 1. string类型字段非必填时必须要在最前面加 "omitempty"
	 * 2. []string 要加 "dive" 才会生效
	 * 3. 有空格的字符串不能使用 "alpha"
	 * 4. []map[string]string 类型需要使用两个 dive 才能控制 key 和 value 的校验规则
	 */

	// 自定义校验规则

	// 校验GET请求的query参数order
	err = validate.Validator.RegisterValidation(
		"order",
		func(fl validator.FieldLevel) bool {
			field := fl.Field()
			switch field.Kind() {
			case reflect.String:
				return regexp.MustCompile(`^([a-zA-Z][A-Za-z0-9_]{0,29}:[ad])(,[a-zA-Z][A-Za-z0-9_]{0,29}:[ad])*$`).MatchString(field.String())
			default:
				return false
			}
		},
	)
	if err != nil {
		return
	}

	return
}
