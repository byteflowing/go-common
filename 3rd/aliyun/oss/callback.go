package oss

import (
	"crypto"
	"crypto/md5"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var ossPublicKeyUrls = []string{
	"https://gosspublic.alicdn.com/",
	"http://gosspublic.alicdn.com/",
}

type ContentType string

const (
	ContentTypeFormUrlEncoded = "application/x-www-form-urlencoded"
	ContentTypeJson           = "application/json"
)

type CallbackParam struct {
	CallbackUrl      string `json:"callbackUrl"`
	CallbackHost     string `json:"callbackHost"`
	CallbackBody     string `json:"callbackBody"`
	CallbackBodyType string `json:"callbackBodyType"`
	CallbackSNI      bool   `json:"callbackSNI"`
}

type CallbackOpts struct {
	// 存储空间名称
	Bucket bool
	// 对象（文件）的完整路径
	Object bool
	// 文件的ETag，即返回给用户的ETag字段
	Etag bool
	// Object大小。
	// 调用CompleteMultipartUpload时，size为整个Object的大小
	Size bool
	// 资源类型，例如jpeg图片的资源类型为image/jpeg
	MimeType bool
	// 与上传文件后返回的x-oss-hash-crc64ecma头内容一致
	CRC64 bool
	// 与上传文件后返回的Content-MD5头内容一致
	// 仅在调用PutObject和PostObject接口上传文件时，该变量的值不为空
	ContentMd5 bool
	// 发起请求的客户端所在的VpcId
	// 如果不是通过VPC发起请求，则该变量的值为空
	VPCId bool
	// 发起请求的客户端IP地址
	ClientIp bool
	// 发起请求的RequestId
	ReqId bool
	// 发起请求的接口名称，例如PutObject、PostObject等
	Operation bool
	// 文件上传成功后，OSS向此URL发送回调请求
	// 请求方法为POST，Body为callbackBody指定的内容。正常情况下，该URL需要响应HTTP/1.1 200 OK，Body必须为JSON格式，响应头Content-Length必须为合法的值，且大小不超过3 MB
	// 支持同时配置最多5个URL，多个URL间以分号（;）分隔。OSS会依次发送请求，直到第一个回调请求成功返回,支持HTTPS地址
	// 为了保证正确处理中文等情况，callbackUrl需做URL编码处理，
	// 例如http://example.com/中文.php?key=value&中文名称=中文值需要编码为
	// http://example.com/%E4%B8%AD%E6%96%87.php?key=value&%E4%B8%AD%E6%96%87%E5%90%8D%E7%A7%B0=%E4%B8%AD%E6%96%87%E5%80%BC
	// e.g. "172.16.XX.XX/test.php"
	CallbackUrl string
	// 发起回调请求时Host头的值，格式为域名或IP地址
	// callbackHost仅在设置了callbackUrl时有效
	// e.g. "oss-cn-hangzhou.aliyuncs.com"
	CallbackHost string
	// 客户端发起回调请求时，OSS是否向通过callbackUrl指定的回源地址发送服务器名称指示SNI（Server Name Indication）
	// 是否发送SNI取决于服务器的配置和需求。对于使用同一个IP地址来托管多个TLS/SSL证书的服务器的情况，建议选择发送SNI。
	CallbackSNI bool
	// 发起回调请求的Content-Type。Content-Type支持以下两种类型：
	//  application/x-www-form-urlencoded: 将经过URL编码的值替换callbackBody中的变量
	//  application/json: 按照JSON格式替换callbackBody中的变量
	CallBackBodyType ContentType
	// 您可以通过callback-var参数来配置自定义参数。
	CallbackVar []string
	// 图片信息
	ImageInfo *ImageInfoOpts
}

type ImageInfoOpts struct {
	// 图片高度。该变量仅适用于图片格式，对于非图片格式，该变量的值为空
	Height bool
	// 图片宽度。该变量仅适用于图片格式，对于非图片格式，该变量的值为空。
	Width bool
	// 图片格式，例如JPG、PNG等。该变量仅适用于图片格式，对于非图片格式，该变量的值为空
	Format bool
}

// NewCallback 仅PutObject、PostObject和CompleteMultipartUpload接口支持设置Callback
// 文档：https://help.aliyun.com/zh/oss/developer-reference/callback?spm=a2c4g.11186623.0.0.5a2e4cf5hV9KdP#ea019ac1e2edt
func NewCallback(opts *CallbackOpts) (callback string, err error) {
	parsedUrl, err := url.Parse(opts.CallbackUrl)
	if err != nil {
		return "", err
	}
	parsedUrl.RawQuery = parsedUrl.Query().Encode()
	ck := &CallbackParam{
		CallbackUrl:      parsedUrl.String(),
		CallbackHost:     opts.CallbackHost,
		CallbackBodyType: string(opts.CallBackBodyType),
		CallbackSNI:      opts.CallbackSNI,
	}
	ck.CallbackBody = generateCallbackBody(opts)
	byteCallback, err := json.Marshal(ck)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(byteCallback), nil
}

func CheckCallback(request *http.Request) (params map[string]interface{}, ok bool, err error) {
	if request.Method != "POST" {
		return nil, false, errors.New("wrong http method: " + request.Method)
	}
	body, err := io.ReadAll(request.Body)
	defer request.Body.Close()
	if err != nil {
		return nil, false, err
	}
	bytePublicKey, err := getPublicKey(request)
	if err != nil {
		return nil, false, err
	}
	byteAuthorization, err := getAuthorization(request)
	if err != nil {
		return nil, false, err
	}
	byteMD5, err := getMD5(request, body)
	if err != nil {
		return nil, false, err
	}
	ok, err = verifySignature(bytePublicKey, byteMD5, byteAuthorization)
	if err != nil {
		return nil, false, err
	}
	params = make(map[string]interface{})
	if err := json.Unmarshal(body, &params); err != nil {
		return nil, false, err
	}
	queries := request.URL.Query()
	for key, values := range queries {
		for _, v := range values {
			params[key] = v
		}
	}
	return params, ok, err
}

func generateCallbackBody(opts *CallbackOpts) string {
	switch opts.CallBackBodyType {
	case ContentTypeJson:
		return generateJsonBody(opts)
	case ContentTypeFormUrlEncoded:
		return generateUrlEncodedBody(opts)
	}
	return ""
}

func parseParams(opts *CallbackOpts) map[string]string {
	params := make(map[string]string)
	if opts.Bucket {
		params["bucket"] = "${bucket}"
	}
	if opts.Object {
		params["object"] = "${object}"
	}
	if opts.Etag {
		params["etag"] = "${etag}"
	}
	if opts.Size {
		params["size"] = "${size}"
	}
	if opts.MimeType {
		params["mimeType"] = "${mimeType}"
	}
	if opts.ImageInfo != nil {
		if opts.ImageInfo.Height {
			params["imageInfo.height"] = "${imageInfo.height}"
		}
		if opts.ImageInfo.Width {
			params["imageInfo.width"] = "${imageInfo.width}"
		}
		if opts.ImageInfo.Format {
			params["imageInfo.format"] = "${imageInfo.format}"
		}
	}
	if opts.CRC64 {
		params["crc64"] = "${crc64}"
	}
	if opts.ContentMd5 {
		params["contentMd5"] = "${contentMd5}"
	}
	if opts.VPCId {
		params["vpcId"] = "${vpcId}"
	}
	if opts.ReqId {
		params["reqId"] = "${reqId}"
	}
	if opts.Operation {
		params["operation"] = "${operation}"
	}
	if len(opts.CallbackVar) > 0 {
		for _, v := range opts.CallbackVar {
			params[v] = fmt.Sprintf("${x:%s}", v)
		}
	}
	return params
}

func generateJsonBody(opts *CallbackOpts) string {
	params := parseParams(opts)
	byteParam, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}
	return string(byteParam)
}

func generateUrlEncodedBody(opts *CallbackOpts) string {
	params := parseParams(opts)
	var sb strings.Builder
	for key, value := range params {
		if sb.Len() > 0 {
			sb.WriteString("&")
		}
		sb.WriteString(fmt.Sprintf("%s=%s", key, value))
	}
	return sb.String()
}

func checkUrl(url string) bool {
	for _, u := range ossPublicKeyUrls {
		if strings.HasPrefix(url, u) {
			return true
		}
	}
	return false
}

func getPublicKey(r *http.Request) ([]byte, error) {
	publicKeyBase64 := r.Header.Get("x-oss-pub-key-url")
	if publicKeyBase64 == "" {
		return nil, errors.New("no x-oss-pub-key-url field found in request header")
	}
	publicKeyURL, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, err
	}
	urlStr := string(publicKeyURL)
	if !checkUrl(urlStr) {
		return nil, fmt.Errorf("unknown oss public key url: %s", urlStr)
	}
	response, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}

func getAuthorization(r *http.Request) ([]byte, error) {
	strAuthorizationBase64 := r.Header.Get("authorization")
	if strAuthorizationBase64 == "" {
		return nil, errors.New("no authorization field found in request header")
	}
	return base64.StdEncoding.DecodeString(strAuthorizationBase64)
}

func getMD5(r *http.Request, body []byte) ([]byte, error) {
	bodyContent := body
	strCallbackBody := string(bodyContent)
	strURLPathDecode, err := unescapePath(r.URL.Path, encodePathSegment)
	if err != nil {
		return nil, err
	}
	var strAuth string
	if r.URL.RawQuery == "" {
		strAuth = fmt.Sprintf("%s\n%s", strURLPathDecode, strCallbackBody)
	} else {
		strAuth = fmt.Sprintf("%s?%s\n%s", strURLPathDecode, r.URL.RawQuery, strCallbackBody)
	}
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(strAuth))
	return md5Ctx.Sum(nil), nil
}

func verifySignature(bytePublicKey []byte, byteMd5 []byte, authorization []byte) (ok bool, err error) {
	pubBlock, _ := pem.Decode(bytePublicKey)
	if pubBlock == nil {
		return false, fmt.Errorf("failed to parse PEM")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		return false, fmt.Errorf("parse pub key error: %s", err.Error())
	}
	if pubInterface == nil {
		return false, fmt.Errorf("x509.ParsePKIXPublicKey failed")
	}
	pub := pubInterface.(*rsa.PublicKey)
	err = rsa.VerifyPKCS1v15(pub, crypto.MD5, byteMd5, authorization)
	if err != nil {
		return false, err
	}
	return true, nil
}

// EscapeError Escape Error
type EscapeError string

func (e EscapeError) Error() string {
	return "invalid URL escape " + strconv.Quote(string(e))
}

// InvalidHostError Invalid Host Error
type InvalidHostError string

func (e InvalidHostError) Error() string {
	return "invalid character " + strconv.Quote(string(e)) + " in host name"
}

type encoding int

const (
	encodePath encoding = 1 + iota
	encodePathSegment
	encodeHost
	encodeZone
	encodeUserPassword
	encodeQueryComponent
	encodeFragment
)

// unescapePath : unescapes a string; the mode specifies, which section of the URL string is being unescaped.
func unescapePath(s string, mode encoding) (string, error) {
	// Count %, check that they're well-formed.
	mode = encodePathSegment
	n := 0
	hasPlus := false
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			n++
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				s = s[i:]
				if len(s) > 3 {
					s = s[:3]
				}
				return "", EscapeError(s)
			}
			// Per https://tools.ietf.org/html/rfc3986#page-21
			// in the host component %-encoding can only be used
			// for non-ASCII bytes.
			// But https://tools.ietf.org/html/rfc6874#section-2
			// introduces %25 being allowed to escape a percent sign
			// in IPv6 scoped-address literals. Yay.
			if mode == encodeHost && unhex(s[i+1]) < 8 && s[i:i+3] != "%25" {
				return "", EscapeError(s[i : i+3])
			}
			if mode == encodeZone {
				// RFC 6874 says basically "anything goes" for zone identifiers
				// and that even non-ASCII can be redundantly escaped,
				// but it seems prudent to restrict %-escaped bytes here to those
				// that are valid host name bytes in their unescaped form.
				// That is, you can use escaping in the zone identifier but not
				// to introduce bytes you couldn't just write directly.
				// But Windows puts spaces here! Yay.
				v := unhex(s[i+1])<<4 | unhex(s[i+2])
				if s[i:i+3] != "%25" && v != ' ' && shouldEscape(v, encodeHost) {
					return "", EscapeError(s[i : i+3])
				}
			}
			i += 3
		case '+':
			hasPlus = mode == encodeQueryComponent
			i++
		default:
			if (mode == encodeHost || mode == encodeZone) && s[i] < 0x80 && shouldEscape(s[i], mode) {
				return "", InvalidHostError(s[i : i+1])
			}
			i++
		}
	}

	if n == 0 && !hasPlus {
		return s, nil
	}

	t := make([]byte, len(s)-2*n)
	j := 0
	for i := 0; i < len(s); {
		switch s[i] {
		case '%':
			t[j] = unhex(s[i+1])<<4 | unhex(s[i+2])
			j++
			i += 3
		case '+':
			if mode == encodeQueryComponent {
				t[j] = ' '
			} else {
				t[j] = '+'
			}
			j++
			i++
		default:
			t[j] = s[i]
			j++
			i++
		}
	}
	return string(t), nil
}

// Please be informed that for now shouldEscape does not check all
// reserved characters correctly. See golang.org/issue/5684.
func shouldEscape(c byte, mode encoding) bool {
	// §2.3 Unreserved characters (alphanum)
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || '0' <= c && c <= '9' {
		return false
	}

	if mode == encodeHost || mode == encodeZone {
		// §3.2.2 Host allows
		//	sub-delims = "!" / "$" / "&" / "'" / "(" / ")" / "*" / "+" / "," / ";" / "="
		// as part of reg-name.
		// We add : because we include :port as part of host.
		// We add [ ] because we include [ipv6]:port as part of host.
		// We add < > because they're the only characters left that
		// we could possibly allow, and Parse will reject them if we
		// escape them (because hosts can't use %-encoding for
		// ASCII bytes).
		switch c {
		case '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=', ':', '[', ']', '<', '>', '"':
			return false
		}
	}

	switch c {
	case '-', '_', '.', '~': // §2.3 Unreserved characters (mark)
		return false

	case '$', '&', '+', ',', '/', ':', ';', '=', '?', '@': // §2.2 Reserved characters (reserved)
		// Different sections of the URL allow a few of
		// the reserved characters to appear unescaped.
		switch mode {
		case encodePath: // §3.3
			// The RFC allows : @ & = + $ but saves / ; , for assigning
			// meaning to individual path segments. This package
			// only manipulates the path as a whole, so we allow those
			// last three as well. That leaves only ? to escape.
			return c == '?'

		case encodePathSegment: // §3.3
			// The RFC allows : @ & = + $ but saves / ; , for assigning
			// meaning to individual path segments.
			return c == '/' || c == ';' || c == ',' || c == '?'

		case encodeUserPassword: // §3.2.1
			// The RFC allows ';', ':', '&', '=', '+', '$', and ',' in
			// userinfo, so we must escape only '@', '/', and '?'.
			// The parsing of userinfo treats ':' as special so we must escape
			// that too.
			return c == '@' || c == '/' || c == '?' || c == ':'

		case encodeQueryComponent: // §3.4
			// The RFC reserves (so we must escape) everything.
			return true

		case encodeFragment: // §4.1
			// The RFC text is silent but the grammar allows
			// everything, so escape nothing.
			return false
		default:
			panic("unhandled default case")
		}
	}

	// Everything else must be escaped.
	return true
}

func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}
