package echos

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	tzh "github.com/go-playground/validator/v10/translations/zh"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Binder struct {
	trans          ut.Translator
	validate       *validator.Validate
	defaultBuilder echo.Binder
}

func NewBinder() (echo.Binder, error) {
	uni := ut.New(zh.New())
	trans, _ := uni.GetTranslator("zh")
	vld, err := NewValidator()
	if err != nil {
		return nil, err
	}
	b := &Binder{
		trans:          trans,
		validate:       vld.Validator,
		defaultBuilder: &echo.DefaultBinder{},
	}
	if err = tzh.RegisterDefaultTranslations(b.validate, b.trans); err != nil {
		return nil, err
	}
	b.validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tag := fld.Tag.Get("json")
		if tag != "" && tag != "-" {
			for _, v := range strings.Split(tag, ",") {
				if v != "omitempty" {
					return v
				}
			}
		}
		return fld.Name
	})
	return b, nil
}

// Bind for bind and validate request parameter
func (s *Binder) Bind(i interface{}, c echo.Context) error {
	if err := s.defaultBuilder.Bind(i, c); err != nil {
		return err
	}
	refValue := reflect.ValueOf(i)
	refKind := refValue.Kind()
	for refKind == reflect.Pointer {
		refValue = refValue.Elem()
		refKind = refValue.Kind()
	}
	if refKind == reflect.Slice {
		for index := 0; index < refValue.Len(); index++ {
			if err := s.validator(refValue.Index(index).Interface()); err != nil {
				return err
			}
		}
	}
	if refKind == reflect.Struct {
		if err := s.validator(i); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("unsupported binding type: %s", reflect.ValueOf(i).Type().String())
}

// validator validate request parameter
func (s *Binder) validator(i interface{}) error {
	refType := reflect.TypeOf(i)
	refKind := refType.Kind()
	for refKind == reflect.Pointer {
		refType = refType.Elem()
		refKind = refType.Kind()
	}
	if refKind != reflect.Struct {
		return nil
	}
	var errs []string
	if err := s.validate.Struct(i); err != nil {
		var ves validator.ValidationErrors
		ok := errors.As(err, &ves)
		if !ok {
			return err
		}
		for _, v := range ves {
			errs = append(errs, v.Translate(s.trans))
		}
	}
	if len(errs) != 0 {
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func example() error {
	e := echo.New()
	// e.Debug = true
	e.HideBanner = true
	e.Validator, _ = NewValidator()
	echoBinder, err := NewBinder()
	if err != nil {
		return err
	}
	e.Binder = echoBinder
	e.Server.ReadTimeout = time.Second * 5
	e.Server.WriteTimeout = time.Second * 10
	e.Server.IdleTimeout = time.Second * 15

	// echo路由注册
	apiV1Group := e.Group("/api/v1")

	// register swagger docs route => /api/v1/swagger/index.html
	apiV1Group.GET("/swagger/*", echoSwagger.WrapHandler)

	middleware := NewMiddleware(nil)
	// ip限流
	apiV1Group.Use(middleware.IpLimiter)

	// 默认的跨域中间件
	apiV1Group.Use(echoMiddleware.CORS())

	// 自定义跨域中间件
	if false {
		apiV1Group.Use(
			echoMiddleware.CORSWithConfig(
				echoMiddleware.CORSConfig{
					Skipper:      echoMiddleware.DefaultSkipper,
					AllowOrigins: []string{"*"},
					AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
					AllowHeaders: []string{"*"},
				},
			),
		)
	}
	return nil
}
