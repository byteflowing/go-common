// Package sms
// 文档：https://help.aliyun.com/zh/sms/developer-reference/api-dysmsapi-2017-05-25-sendsms?spm=a2c4g.11186623.0.0.31ba614c4SMnC0
package sms

import (
	"errors"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	smsCli "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	smsv1 "github.com/byteflowing/go-common/gen/sms/v1"
	"github.com/byteflowing/go-common/jsonx"
)

type Sms struct {
	accessKeyId     string
	accessKeySecret string
	securityToken   *string
	cli             *smsCli.Client
}

func New(opts *smsv1.SmsProvider) (s *Sms, err error) {
	smsClient, err := smsCli.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(opts.AccessKey),
		AccessKeySecret: tea.String(opts.SecretKey),
		SecurityToken:   opts.SecurityToken,
	})
	return &Sms{
		accessKeyId:     opts.AccessKey,
		accessKeySecret: opts.SecretKey,
		securityToken:   opts.SecurityToken,
		cli:             smsClient,
	}, nil
}

func (s *Sms) SendSms(req *smsv1.SendSmsReq) (resp *smsv1.SendSmsResp, err error) {
	var params string
	if len(req.TemplateParams) > 0 {
		params, err = jsonx.MarshalToString(req.TemplateParams)
		if err != nil {
			return nil, err
		}
	}
	request := &smsCli.SendSmsRequest{
		PhoneNumbers: tea.String(req.PhoneNumber.Number),
		SignName:     tea.String(req.SignName),
		TemplateCode: tea.String(req.TemplateCode),
	}
	if len(params) > 0 {
		request.TemplateParam = tea.String(params)
	}
	res, err := s.cli.SendSms(request)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, errors.New("response is nil")
	}
	err = s.parseErr(res.Body.BizId, res.Body.RequestId, res.Body.Code, res.Body.Message)
	if err != nil {
		return nil, err
	}
	resp = &smsv1.SendSmsResp{
		ErrMsg: "OK",
	}
	return
}

func (s *Sms) parseErr(bizID, requestID, errCode, errMsg *string) (err error) {
	_bizID := tea.StringValue(bizID)
	_requestID := tea.StringValue(requestID)
	_errCode := tea.StringValue(errCode)
	_errMsg := tea.StringValue(errMsg)
	if _errCode != "OK" {
		return fmt.Errorf("[bizID:%s, requestID:%s, code:%s] errMsg:%s", _bizID, _requestID, _errCode, _errMsg)
	}
	return nil
}
