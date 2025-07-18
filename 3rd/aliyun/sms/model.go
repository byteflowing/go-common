package sms

type SendStatus int

const (
	SendStatusWait = iota + 1
	SendStatusFailed
	SendStatusSuccess
)

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

type QuerySendDetailReq struct {
	Phone string // 手机号码
	BizId string // 发送回执
	Data  string // 发送日期 yyyyMMdd
}

type QuerySendDetailResp struct {
	Common       *CommonResp
	ErrCode      string
	TemplateCode string
	ReceiveDate  string
	SendDate     string
	Phone        string
	Content      string
	Status       SendStatus
}
