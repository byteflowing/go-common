package idx

import (
	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/bytedance/gopkg/lang/stringx"
)

type CaptchaType string

const (
	CaptchaNumber                CaptchaType = "number" // 存数字
	CaptchaLetter                CaptchaType = "letter" //
	CaptchaNumberLetter          CaptchaType = "number_letter"
	CaptchaLetterUppercase       CaptchaType = "letter_uppercase"
	CaptchaLetterLowercase       CaptchaType = "letter_lowercase"
	CaptchaNumberLetterUppercase CaptchaType = "number_letter_uppercase"
	CaptchaNumberLetterLowercase CaptchaType = "number_letter_lowercase"
)

var (
	numberCharset          = "0123456789"
	letterLowercaseCharset = "abcdefghijklmnopqrstuvwxyz"
	letterUppercaseCharset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func GenerateCaptcha(length int, captchaType CaptchaType) string {
	var charset string
	switch captchaType {
	case CaptchaNumber:
		charset = numberCharset
	case CaptchaLetter:
		charset = letterUppercaseCharset + letterLowercaseCharset
	case CaptchaNumberLetter:
		charset = numberCharset + letterLowercaseCharset + letterUppercaseCharset
	case CaptchaLetterUppercase:
		charset = letterUppercaseCharset
	case CaptchaLetterLowercase:
		charset = letterLowercaseCharset
	case CaptchaNumberLetterUppercase:
		charset = numberCharset + letterUppercaseCharset
	case CaptchaNumberLetterLowercase:
		charset = numberCharset + letterLowercaseCharset
	default:
		charset = numberCharset
	}
	charset = stringx.Shuffle(charset)
	captcha := make([]byte, length)
	for i := range captcha {
		captcha[i] = charset[fastrand.Intn(len(charset))]
	}
	return string(captcha)
}
