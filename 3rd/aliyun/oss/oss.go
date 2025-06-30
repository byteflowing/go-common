package oss

import (
	"net/http"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Oss struct {
	accessKeyId     string
	accessKeySecret string
	endPoint        string
	regionId        string
	securityToken   *string

	*oss.Client
}

type Opts struct {
	AccessKeyId     string
	AccessKeySecret string
	EndPoint        string
	RegionId        string
	// [可选] 如果通过sts获取的accessKeyId和AccessKeySecret需要提供此值
	SecurityToken *string
}

type PostPolicy struct {
	Expiration string        `json:"expiration"`
	Conditions []interface{} `json:"conditions"`
}

type GetPostPolicyReq struct {
	// 指定过期时间，单位为秒
	ExpiredTime int64
	// Bucket名称
	Bucket *string
	// 文件最小限制（单位：字节）
	MinLength *int64
	// 文件最大限制（单位：字节）
	MaxLength *int64
	// 指定上传到OSS的文件前缀
	// 例如："user-dir-prefix/"
	UploadDir *string
	// MIME
	// 例如： "image/jpg", "image/png"
	ContentTypes []string
	// 其他conditions
	// conditions参考下面链接
	// https://help.aliyun.com/zh/oss/developer-reference/signature-version-4-recommend?spm=a2c4g.11186623.0.0.4ed44c47q7mV24
	Conditions []interface{}
	// callback选项，如果不指定则不需要回调
	Callback *CallbackOpts
}

type GetPolicyTokenResp struct {
	Policy           string `json:"policy"`
	Callback         string `json:"callback"`
	SignatureVersion string `json:"signature-version"`
	Credential       string `json:"credential"`
	Date             string `json:"date"`
	Signature        string `json:"signature"`
	ExpiredTime      string `json:"expired-time"`
	SecurityToken    string `json:"security-token"`
}

func New(opts *Opts) (o *Oss, err error) {
	var ossOptions []oss.ClientOption
	if opts.SecurityToken != nil {
		ossOptions = append(ossOptions, oss.SecurityToken(*opts.SecurityToken))
	}
	ossClient, err := oss.New(opts.EndPoint, opts.AccessKeyId, opts.AccessKeySecret, ossOptions...)
	if err != nil {
		return nil, err
	}
	return &Oss{
		accessKeyId:     opts.AccessKeyId,
		accessKeySecret: opts.AccessKeySecret,
		endPoint:        opts.EndPoint,
		regionId:        opts.RegionId,
		Client:          ossClient,
		securityToken:   opts.SecurityToken,
	}, nil
}

// GetPostPolicy 获取postObject Policy
// 文档：https://help.aliyun.com/zh/oss/use-cases/uploading-objects-to-oss-directly-from-clients/?spm=a2c4g.11186623.0.0.78495d03iHTHaV#36c322a437r3k
// v4签名：https://help.aliyun.com/zh/oss/developer-reference/signature-version-4-recommend?spm=a2c4g.11186623.0.0.4b2e5f92PAA5b8
// postObject: https://help.aliyun.com/zh/oss/developer-reference/postobject?spm=a2c4g.11186623.0.0.5e475f92qkDehJ#section-0lu-03w-2z6
func (o *Oss) GetPostPolicy(req *GetPostPolicyReq) (resp *GetPolicyTokenResp, err error) {
	return o.getPolicyToken(req)
}

// GetCallback 仅PutObject、PostObject和CompleteMultipartUpload接口支持设置Callback
// 文档：https://help.aliyun.com/zh/oss/developer-reference/callback?spm=a2c4g.11186623.0.0.5a2e4cf5hV9KdP#ea019ac1e2edt
func (o *Oss) GetCallback(opts *CallbackOpts) (callback string, err error) {
	return NewCallback(opts)
}

// CheckCallback 校验callback请求是否合法
// @Returns params为从callback请求中提取的回调参数
func (o *Oss) CheckCallback(r *http.Request) (params map[string]interface{}, ok bool, err error) {
	return CheckCallback(r)
}
