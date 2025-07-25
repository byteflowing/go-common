// Package sms
// 文档：https://help.aliyun.com/zh/sms/developer-reference/api-dysmsapi-2017-05-25-sendsms?spm=a2c4g.11186623.0.0.31ba614c4SMnC0
package sms

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	smsCli "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/byteflowing/go-common/jsonx"
)

type Opts struct {
	AccessKeyId     string
	AccessKeySecret string
	// [可选] 如果通过sts获取的accessKeyId和AccessKeySecret需要提供此值
	SecurityToken *string
}

type Sms struct {
	accessKeyId     string
	accessKeySecret string
	securityToken   *string
	cli             *smsCli.Client
}

func New(opts *Opts) (s *Sms, err error) {
	smsClient, err := smsCli.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(opts.AccessKeyId),
		AccessKeySecret: tea.String(opts.AccessKeySecret),
		SecurityToken:   opts.SecurityToken,
	})
	return &Sms{
		accessKeyId:     opts.AccessKeyId,
		accessKeySecret: opts.AccessKeySecret,
		securityToken:   opts.SecurityToken,
		cli:             smsClient,
	}, nil
}

func (s *Sms) SendSms(req *SendSmsReq) (resp *SendSmsResp, err error) {
	var params string
	if len(req.TemplateParam) > 0 {
		params, err = jsonx.MarshalToString(req.TemplateParam)
		if err != nil {
			return nil, err
		}
	}
	request := &smsCli.SendSmsRequest{
		PhoneNumbers: tea.String(req.PhoneNumbers),
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
	resp = &SendSmsResp{
		Common: &CommonResp{
			BizId:     tea.StringValue(res.Body.BizId),
			Code:      tea.StringValue(res.Body.Code),
			Message:   tea.StringValue(res.Body.Message),
			RequestId: tea.StringValue(res.Body.RequestId),
		}}
	return
}

func (s *Sms) QuerySendDetail(req *QuerySendDetailReq) (resp *QuerySendDetailResp, err error) {
	result, err := s.cli.QuerySendDetails(&smsCli.QuerySendDetailsRequest{
		BizId:       tea.String(req.BizId),
		CurrentPage: tea.Int64(1),
		PageSize:    tea.Int64(1),
		PhoneNumber: tea.String(req.Phone),
		SendDate:    tea.String(req.Data),
	})
	if err != nil {
		return nil, err
	}
	resp = &QuerySendDetailResp{
		Common: &CommonResp{
			Code:      tea.StringValue(result.Body.Code),
			Message:   tea.StringValue(result.Body.Message),
			RequestId: tea.StringValue(result.Body.RequestId),
		},
	}
	if result.Body.SmsSendDetailDTOs != nil && len(result.Body.SmsSendDetailDTOs.SmsSendDetailDTO) > 0 {
		v := result.Body.SmsSendDetailDTOs.SmsSendDetailDTO[0]
		resp.ErrCode = v.ErrCode
		resp.TemplateCode = v.TemplateCode
		resp.ReceiveDate = v.ReceiveDate
		resp.SendDate = v.SendDate
		resp.Phone = v.PhoneNum
		resp.Content = v.PhoneNum
		resp.Status = s.parseStatus(v.SendStatus)
	}
	return
}

func (s *Sms) parseStatus(status *int64) SendStatus {
	v := tea.Int64Value(status)
	var st SendStatus
	switch v {
	case 1:
		st = SendStatusWait
	case 2:
		st = SendStatusFailed
	case 3:
		st = SendStatusSuccess
	default:
		st = SendStatusFailed
	}
	return st
}
