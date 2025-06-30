package sms

type CommonResp struct {
	BizId     string
	Code      string
	Message   string
	RequestId string
}

type SendSmsReq struct {
	PhoneNumbers  string
	SignName      string
	TemplateCode  string
	TemplateParam map[string]string
}

type SendSmsResp struct {
	Common *CommonResp
}
