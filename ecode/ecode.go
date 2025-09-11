package ecode

import (
	"errors"
	"fmt"
)

var (
	_codes  = make(map[uint32]*Code)
	_errMsg = make(map[string]*Code)
)

type Codes interface {
	Error() string       // 返回详细错误信息
	Code() uint32        // 返回错误码
	Message() string     // 返回错误的说明信息
	AddDetail(err error) // 添加底层错误信息
}

type Code struct {
	ErrCode uint32 // 错误代码
	ErrMsg  string // 错误信息
	Detail  error  // 保存内部错误，主要用于日志，不向前端暴露
}

// NewCode : 非业务错误码小于0， 业务代码大于0
func NewCode(errCode uint32, errMsg string) Codes {
	return add(errCode, errMsg)
}

func (c *Code) Message() string {
	if len(c.ErrMsg) > 0 {
		return c.ErrMsg
	}
	return _codes[c.ErrCode].ErrMsg
}

func (c *Code) Code() uint32 {
	return c.ErrCode
}

func (c *Code) AddDetail(err error) {
	if c.Detail == nil {
		c.Detail = err
	} else {
		c.Detail = fmt.Errorf("%w; %s", c.Detail, err.Error())
	}
}

func (c *Code) Error() string {
	if c.Detail != nil {
		return fmt.Sprintf("code: %d, message: %s, detail: %s", c.ErrCode, c.Message(), c.Detail.Error())
	}
	return fmt.Sprintf("code: %d, message: %s", c.ErrCode, c.Message())
}

func IsCodes(err error) (Codes, bool) {
	if err == nil {
		return nil, false
	}
	var c Codes
	if errors.As(err, &c) {
		return c, true
	}
	return nil, false
}

func FromCode(code uint32) Codes {
	c, ok := _codes[code]
	if !ok {
		return nil
	}
	return c
}

func FromMsg(msg string) Codes {
	m, ok := _errMsg[msg]
	if !ok {
		return nil
	}
	return m
}

func add(errCode uint32, errMsg string) *Code {
	if _, ok := _codes[errCode]; ok {
		panic(fmt.Sprintf("code: %d already exist", errCode))
	}
	if _, ok := _errMsg[errMsg]; ok {
		panic(fmt.Sprintf("message: %s already exist", errMsg))
	}
	code := &Code{
		ErrCode: errCode,
		ErrMsg:  errMsg,
	}
	_codes[errCode] = code
	_errMsg[errMsg] = code
	return code
}
