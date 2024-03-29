package echos

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

const (
	CodeSuccess  = 0 // 处理成功
	CodeFail     = 1 // 处理失败
	CodeError    = 2 // 服务器出错
	CodeAbnormal = 3 // 其它业务异常
	CodeBad      = 4 // 客户端参数异常
)

const (
	MsgSuccess = "SUCCESS"
)

type Resp struct {
	Code  int         `json:"code" xml:"code" validate:"required,min=0"`                        // 业务状态码
	Msg   string      `json:"msg" xml:"msg" validate:"required"`                                // 业务描述语
	Bag   interface{} `json:"bag,omitempty" xml:"bag,omitempty" validate:"omitempty"`           // 业务数据包
	Count *int64      `json:"count,omitempty" xml:"count,omitempty" validate:"omitempty,min=0"` // 业务数据总条数
}

type Ctx struct {
	echo.Context

	ReqId string // request id

	Uid    int64  // store account id for int64 type
	UidStr string // store account id for string type

	status int // http status code

	resp *Resp // response content
}

func NewCtx() *Ctx {
	return &Ctx{resp: &Resp{}}
}

// clean clear the property values of the current object
func (s *Ctx) clean() {
	s.Context = nil
	s.ReqId = ""
	s.Uid = 0
	s.UidStr = ""
	s.status = 0
	s.resp.Code = 0
	s.resp.Msg = ""
	s.resp.Bag = nil
	s.resp.Count = nil
}

func (s *Ctx) Status(status int) *Ctx {
	s.status = status
	return s
}

func (s *Ctx) Code(cod int) *Ctx {
	s.resp.Code = cod
	return s
}

func (s *Ctx) Msg(msg string) *Ctx {
	s.resp.Msg = msg
	return s
}

func (s *Ctx) Bag(bag interface{}) *Ctx {
	s.resp.Bag = bag
	return s
}

func (s *Ctx) Count(count int64) *Ctx {
	s.resp.Count = &count
	return s
}

func (s *Ctx) Json() error {
	defer s.clean()
	return s.JSON(s.status, s.resp)
}

func (s *Ctx) Ok() error {
	return s.Status(http.StatusOK).Code(CodeSuccess).Msg(MsgSuccess).Json()
}

func (s *Ctx) Abn(err error) error {
	msg := err.Error()
	if logger := s.Logger(); logger != nil {
		logger.Warnf("%s --> %s", s.Context.Path(), msg)
	}
	return s.Status(http.StatusOK).Code(CodeAbnormal).Msg(msg).Json()
}

func (s *Ctx) Bad(err error) error {
	msg := err.Error()
	if logger := s.Logger(); logger != nil {
		logger.Warnf("%s --> %s", s.Context.Path(), msg)
	}
	return s.Status(http.StatusOK).Code(CodeBad).Msg(msg).Json()
}

func (s *Ctx) Err(err error) error {
	if logger := s.Logger(); logger != nil {
		logger.Errorf("%s --> %s", s.Context.Path(), err.Error())
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
