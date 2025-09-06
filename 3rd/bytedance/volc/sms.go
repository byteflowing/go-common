package volc

import (
	"fmt"

	smsv1 "github.com/byteflowing/go-common/gen/sms/v1"
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

func NewSms(opts *smsv1.SmsProvider) *Sms {
	cli := sms.DefaultInstance
	cli.Client.SetAccessKey(opts.AccessKey)
	cli.Client.SetSecretKey(opts.SecretKey)
	return &Sms{
		accessKeyId:     opts.AccessKey,
		accessKeySecret: opts.SecretKey,
		cli:             cli,
	}
}

func (s *Sms) SendSms(req *smsv1.SendSmsReq) (resp *smsv1.SendSmsResp, err error) {
	params, err := jsonx.MarshalToString(req.TemplateParams)
	if err != nil {
		return nil, err
	}
	res, _, err := s.cli.Send(&sms.SmsRequest{
		SmsAccount:    req.Account,
		Sign:          req.SignName,
		TemplateID:    req.TemplateCode,
		TemplateParam: params,
		PhoneNumbers:  req.PhoneNumber.Number,
	})
	err = s.parseErr(res, err)
	if err != nil {
		return nil, err
	}
	resp = &smsv1.SendSmsResp{
		ErrMsg: "OK",
	}
	return
}

func (s *Sms) parseErr(resp *sms.SmsResponse, err error) error {
	if err != nil {
		if resp != nil {
			return fmt.Errorf(
				"requestID: %s, action: %s, version: %s, service: %s, region: %s, messageID: %v, err: %v",
				resp.ResponseMetadata.RequestId,
				resp.ResponseMetadata.Action,
				resp.ResponseMetadata.Version,
				resp.ResponseMetadata.Service,
				resp.ResponseMetadata.Region,
				resp.Result.MessageID,
				err,
			)
		}
		return err
	}
	return nil
}
