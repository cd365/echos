package echos

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	CodeSuccess = 0
	CodeFail    = 1
	CodeError   = 2
)

const (
	MsgSuccess = "SUCCESS"
)

type Ctx struct {
	echo.Context `json:"-" xml:"-" validate:"-"`

	logger echo.Logger

	ReqId string `json:"-" xml:"-" validate:"-"` // request id

	Uid    int64  `json:"-" xml:"-" validate:"-"` // store account id for int64 type
	UidStr string `json:"-" xml:"-" validate:"-"` // store account id for string type

	status int // http status code

	OfCode  int         `json:"code" xml:"code" validate:"required,min=0"`                        // 业务状态码
	OfMsg   string      `json:"msg" xml:"msg" validate:"required"`                                // 业务描述语
	OfBag   interface{} `json:"bag,omitempty" xml:"bag,omitempty" validate:"omitempty"`           // 业务数据包
	OfCount *int64      `json:"count,omitempty" xml:"count,omitempty" validate:"omitempty,min=0"` // 业务数据总条数
}

func NewCtx(
	context echo.Context,
	logger echo.Logger,
) *Ctx {
	return &Ctx{
		Context: context,
		logger:  logger,
	}
}

func (s *Ctx) Status(status int) *Ctx {
	s.status = status
	return s
}

func (s *Ctx) Code(cod int) *Ctx {
	s.OfCode = cod
	return s
}

func (s *Ctx) Msg(msg string) *Ctx {
	s.OfMsg = msg
	return s
}

func (s *Ctx) Bag(bag interface{}) *Ctx {
	s.OfBag = bag
	return s
}

func (s *Ctx) Count(count int64) *Ctx {
	s.OfCount = &count
	return s
}

func (s *Ctx) Json() error {
	return s.JSON(s.status, s)
}

func (s *Ctx) Ok() error {
	return s.Status(http.StatusOK).Code(CodeSuccess).Msg(MsgSuccess).Json()
}

func (s *Ctx) Bad(err error) error {
	return s.Status(http.StatusOK).Code(400000).Msg(err.Error()).Json()
}

func (s *Ctx) Err(err error) error {
	if s.logger != nil && err != nil {
		s.logger.Error(err.Error())
	}
	return s.Status(http.StatusInternalServerError).Code(CodeError).Msg(http.StatusText(http.StatusInternalServerError)).Json()
}

func (s *Ctx) Fail(msg string) error {
	return s.Status(http.StatusOK).Code(CodeFail).Msg(msg).Json()
}

func (s *Ctx) Data(bag interface{}, count ...int64) error {
	s.Status(http.StatusOK).Code(CodeSuccess).Msg(MsgSuccess).Bag(bag)
	if length := len(count); length > 0 {
		s.Count(count[length-1])
	}
	return s.Json()
}
