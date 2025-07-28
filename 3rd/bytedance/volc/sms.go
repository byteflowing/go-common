package volc

import (
	"fmt"
	"strings"

	"github.com/byteflowing/go-common/jsonx"
	"github.com/volcengine/volc-sdk-golang/service/sms"
)

// 文档：https://www.volcengine.com/docs/6361/67380
// go sdk：https://www.volcengine.com/docs/6361/1109261

type Sms struct {
	accessKeyId     string
	accessKeySecret string
	cli             *sms.SMS
}

type SmsOpts struct {
	AccessKeyId     string
	AccessKeySecret string
}

func NewSms(opts *SmsOpts) *Sms {
	cli := sms.DefaultInstance
	cli.Client.SetAccessKey(opts.AccessKeyId)
	cli.Client.SetSecretKey(opts.AccessKeySecret)
	return &Sms{
		accessKeyId:     opts.AccessKeyId,
		accessKeySecret: opts.AccessKeySecret,
		cli:             cli,
	}
}

type SmsCommonResp struct {
	RequestId string
	Action    string
	Version   string
	Service   string
	Region    string
	ErrCode   string
	ErrMsg    string
}

type SendSmsReq struct {
	SmsAccount    string
	Sign          string
	TemplateID    string
	TemplateParam map[string]string
	PhoneNumber   string
}

type SendSmsResp struct {
	*SmsCommonResp
	MessageID string
}

type SendSmsToMultiPhoneReq struct {
	SmsAccount    string
	Sign          string
	TemplateID    string
	TemplateParam map[string]string
	PhoneNumbers  []string
}

type SendSmsToMultiPhoneResp struct {
	*SmsCommonResp
	MessageID []string
}

func (s *Sms) SendSms(req *SendSmsReq) (resp *SendSmsResp, err error) {
	params, err := jsonx.MarshalToString(req.TemplateParam)
	if err != nil {
		return nil, err
	}
	res, _, err := s.cli.Send(&sms.SmsRequest{
		SmsAccount:    req.SmsAccount,
		Sign:          req.Sign,
		TemplateID:    req.TemplateID,
		TemplateParam: params,
		PhoneNumbers:  req.PhoneNumber,
	})
	if err != nil {
		return nil, err
	}
	commonResp := &SmsCommonResp{
		RequestId: res.ResponseMetadata.RequestId,
		Action:    res.ResponseMetadata.Action,
		Version:   res.ResponseMetadata.Version,
		Service:   res.ResponseMetadata.Service,
		Region:    res.ResponseMetadata.Region,
	}
	if res.ResponseMetadata.Error != nil {
		commonResp.ErrCode = res.ResponseMetadata.Error.Code
		commonResp.ErrMsg = res.ResponseMetadata.Error.Message
	}
	err = fmt.Errorf("errCode: %s, errMsg: %s", commonResp.ErrCode, commonResp.ErrMsg)
	resp = &SendSmsResp{
		SmsCommonResp: commonResp,
		MessageID:     res.Result.MessageID[0],
	}
	return
}

func (s *Sms) SendSmsToMultiPhone(req *SendSmsToMultiPhoneReq) (resp *SendSmsToMultiPhoneResp, err error) {
	phones := strings.Join(req.PhoneNumbers, ",")
	params, err := jsonx.MarshalToString(req.TemplateParam)
	if err != nil {
		return nil, err
	}
	res, _, err := s.cli.Send(&sms.SmsRequest{
		SmsAccount:    req.SmsAccount,
		Sign:          req.Sign,
		TemplateID:    req.TemplateID,
		TemplateParam: params,
		PhoneNumbers:  phones,
	})
	if err != nil {
		return nil, err
	}
	commonResp := &SmsCommonResp{
		RequestId: res.ResponseMetadata.RequestId,
		Action:    res.ResponseMetadata.Action,
		Version:   res.ResponseMetadata.Version,
		Service:   res.ResponseMetadata.Service,
		Region:    res.ResponseMetadata.Region,
	}
	if res.ResponseMetadata.Error != nil {
		commonResp.ErrCode = res.ResponseMetadata.Error.Code
		commonResp.ErrMsg = res.ResponseMetadata.Error.Message
	}
	err = fmt.Errorf("errCode: %s, errMsg: %s", commonResp.ErrCode, commonResp.ErrMsg)
	resp = &SendSmsToMultiPhoneResp{
		SmsCommonResp: commonResp,
		MessageID:     res.Result.MessageID,
	}
	return
}
