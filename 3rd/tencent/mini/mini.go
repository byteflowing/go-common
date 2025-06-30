package mini

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/byteflowing/go-common/jsonx"
	"io"
	"net/http"
	"net/url"
)

const (
	code2SessionURL    = "https://api.weixin.qq.com/sns/jscode2session"
	checkSessionKeyURl = "https://api.weixin.qq.com/wxa/checksession"
	getAccessTokenURL  = "https://api.weixin.qq.com/cgi-bin/token"
	resetSessionKeyURL = "https://api.weixin.qq.com/wxa/resetusersessionkey"

	loginGrantType       = "authorization_code"
	accessTokenGrantType = "client_credential"

	sigMethodHMACSHA256 = "hmac_sha256"
)

type Client struct {
	AppID      string // 小程序 appId
	Secret     string // 小程序 appSecret
	httpClient *http.Client
}

type Opts struct {
	AppID  string // 小程序 appId
	Secret string // 小程序 appSecret
}

func NewMiniClient(opts *Opts) *Client {
	return &Client{
		AppID:      opts.AppID,
		Secret:     opts.Secret,
		httpClient: http.DefaultClient,
	}
}

func (mini *Client) request(reqURL string) ([]byte, error) {
	response, err := mini.httpClient.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}

// WechatLogin 小程序登录
// 文档：https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-login/code2Session.html
func (mini *Client) WechatLogin(ctx context.Context, req *WechatLoginReq) (resp *WechatLoginResp, err error) {
	params := url.Values{}
	params.Add("js_code", req.Code)
	params.Add("appid", mini.AppID)
	params.Add("secret", mini.Secret)
	params.Add("grant_type", loginGrantType)
	reqURL := fmt.Sprintf("%s?%s", code2SessionURL, params.Encode())
	body, err := mini.request(reqURL)
	if err != nil {
		return nil, err
	}
	resp = &WechatLoginResp{}
	if err = jsonx.Unmarshal(body, resp); err != nil {
		return nil, err
	}
	if err = mini.checkWechatErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return nil, err
	}
	return
}

// GetAccessToken : 获取接口调用凭据
// 文档：https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-access-token/getAccessToken.html
func (mini *Client) GetAccessToken(ctx context.Context, _ *GetAccessTokenReq) (resp *GetAccessTokenResp, err error) {
	params := url.Values{}
	params.Add("appid", mini.AppID)
	params.Add("secret", mini.Secret)
	params.Add("grant_type", accessTokenGrantType)
	reqURL := fmt.Sprintf("%s?%s", getAccessTokenURL, params.Encode())
	body, err := mini.request(reqURL)
	if err != nil {
		return nil, err
	}
	resp = &GetAccessTokenResp{}
	if err = jsonx.Unmarshal(body, resp); err != nil {
		return nil, err
	}
	if err = mini.checkWechatErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return nil, err
	}
	return
}

// CheckLoginStatus : 检测登录状态
// 文档: https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-login/checkSessionKey.html
func (mini *Client) CheckLoginStatus(ctx context.Context, accessToken, sessionKey, openID string) (ok bool, err error) {
	signature := mini.signSessionKey(sessionKey)
	params := url.Values{}
	params.Add("openid", openID)
	params.Add("access_token", accessToken)
	params.Add("signature", signature)
	params.Add("sig_method", sigMethodHMACSHA256)
	reqURL := fmt.Sprintf("%s?%s", checkSessionKeyURl, params.Encode())
	body, err := mini.request(reqURL)
	if err != nil {
		return false, err
	}
	resp := &CommonResp{}
	if err = jsonx.Unmarshal(body, resp); err != nil {
		return false, err
	}
	if resp.ErrCode != 0 {
		return false, nil
	}
	return true, nil
}

// ResetSessionKey 重置登录态
// 文档：https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/user-login/ResetUserSessionKey.html
func (mini *Client) ResetSessionKey(ctx context.Context, req *ResetSessionReq) (resp *ResetSessionResp, err error) {
	signature := mini.signSessionKey(req.SessionKey)
	params := url.Values{}
	params.Add("access_token", req.AccessToken)
	params.Add("openid", req.OpenID)
	params.Add("signature", signature)
	params.Add("sig_method", sigMethodHMACSHA256)
	reqURL := fmt.Sprintf("%s?%s", resetSessionKeyURL, params.Encode())
	body, err := mini.request(reqURL)
	if err != nil {
		return nil, err
	}
	resp = &ResetSessionResp{}
	if err = jsonx.Unmarshal(body, resp); err != nil {
		return nil, err
	}
	if err = mini.checkWechatErr(resp.ErrCode, resp.ErrMsg); err != nil {
		return nil, err
	}
	return
}

func (mini *Client) signSessionKey(sessionKey string) (signature string) {
	h := hmac.New(sha256.New, []byte(sessionKey))
	h.Write([]byte(sessionKey))
	return hex.EncodeToString(h.Sum(nil))
}

func (mini *Client) checkWechatErr(errCode int, errMsg string) (err error) {
	if errCode == 0 {
		return nil
	}
	return fmt.Errorf("%d:%s", errCode, errMsg)
}
