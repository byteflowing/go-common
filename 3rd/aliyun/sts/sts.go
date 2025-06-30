package sts

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sts "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/byteflowing/go-common/jsonx"
)

// STS
// 接入点：https://help.aliyun.com/zh/ram/developer-reference/api-sts-2015-04-01-endpoint?spm=a2c4g.11186623.0.0.530165d9o58kgG#main-107864
// 客户端直传示例：https://help.aliyun.com/zh/oss/use-cases/uploading-objects-to-oss-directly-from-clients/?spm=a2c4g.11186623.0.0.78495d03iHTHaV#78a72119a79q5
// sts文档：https://help.aliyun.com/zh/ram/developer-reference/sts-sdk-overview?spm=a2c4g.11186623.0.0.c242af15zLi10w
// 参数说明：https://help.aliyun.com/zh/ram/developer-reference/api-sts-2015-04-01-assumerole?spm=a2c4g.11186623.0.0.6fb134bdMs0196
// 注意事项：
// STS Token 自颁发后将在一段时间内有效，建议您设置合理的 Token 有效期，并在有效期内重复使用，以避免业务请求速率上升后，STS Token 颁发的速率限制影响到业务。
// 具体速率限制，请参见 STS 服务调用次数是否有上限。您可以通过请求参数DurationSeconds设置 Token 有效期。
// 在移动端上传或下载 OSS 文件等场景下，其访问量较大，即使重复使用 STS Token 也可能无法满足限流要求。为避免 STS 的限流成为 OSS 访问量的瓶颈，您可以尝试 OSS 的在 URL 中包含签名的方案。
// 在URL中包含签名方案：https://help.aliyun.com/zh/oss/developer-reference/ddd-signatures-to-urls?spm=a2c4g.11186623.0.0.72101dd4lfsoDB
// 服务端签名直传方案：https://help.aliyun.com/zh/oss/use-cases/obtain-signature-information-from-the-server-and-upload-data-to-oss?spm=a2c4g.11186623.0.0.72101dd4WSibgp
type STS struct {
	config *Opts
	cli    *sts.Client
}

type Opts struct {
	AccessKeyId     string
	AccessKeySecret string
	RegionId        string
	Endpoint        string
	ConnectTimeout  *int
	ReadTimeout     *int
}

func New(opts *Opts) (client *STS, err error) {
	stsClient, err := sts.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(opts.AccessKeyId),
		AccessKeySecret: tea.String(opts.AccessKeySecret),
		RegionId:        tea.String(opts.RegionId),
		Endpoint:        tea.String(opts.Endpoint),
		ReadTimeout:     opts.ReadTimeout,
		ConnectTimeout:  opts.ConnectTimeout,
	})
	if err != nil {
		return nil, err
	}
	return &STS{
		cli:    stsClient,
		config: opts,
	}, nil
}

// AssumeRole 获取sts
// Policy文档：https://help.aliyun.com/zh/ram/user-guide/policy-elements?spm=a2c4g.11186623.0.0.72104676oFzZr5
// Policy实例：https://help.aliyun.com/zh/ram/user-guide/example-policies/?spm=a2c4g.11186623.0.i9
func (auth *STS) AssumeRole(req *AssumeRoleReq) (resp *AssumeRoleResp, err error) {
	request := &sts.AssumeRoleRequest{
		RoleArn:         tea.String(req.RoleArn),
		DurationSeconds: req.DurationSeconds,
		ExternalId:      req.ExternalId,
		RoleSessionName: req.RoleSessionName,
	}
	if req.Policy != nil {
		policyByte, err := jsonx.Marshal(req.Policy)
		if err != nil {
			return nil, err
		}
		policy := string(policyByte)
		request.Policy = tea.String(policy)
	}
	res, err := auth.cli.AssumeRole(request)
	if err != nil {
		return nil, err
	}
	return &AssumeRoleResp{
		Common:          &CommonResp{RequestId: tea.StringValue(res.Body.RequestId)},
		AccessKeyId:     tea.StringValue(res.Body.Credentials.AccessKeyId),
		AccessKeySecret: tea.StringValue(res.Body.Credentials.AccessKeySecret),
		SecurityToken:   tea.StringValue(res.Body.Credentials.SecurityToken),
		Expiration:      tea.StringValue(res.Body.Credentials.Expiration),
	}, nil
}
