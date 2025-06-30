package mini

type CommonResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type WechatLoginReq struct {
	Code string
}

type WechatLoginResp struct {
	CommonResp // 调用者不用关心这个结构
	OpenID     string
	SessionKey string
	UnionID    string
}

type GetAccessTokenReq struct{}

type GetAccessTokenResp struct {
	CommonResp         // 调用者不用关心这个结构
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type ResetSessionReq struct {
	AccessToken string
	OpenID      string
	SessionKey  string
}

type ResetSessionResp struct {
	CommonResp        // 调用者不用关心这个结构
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
}
