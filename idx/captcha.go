package idx

import (
	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/bytedance/gopkg/lang/stringx"
)

type CaptchaType uint8

const (
	CaptchaNumber CaptchaType = 1 << iota // 存数字
	CaptchaUppercase
	CaptchaLowercase
	CaptchaSymbol
)

var (
	numberCharset = "0123456789"
	lowerCharset  = "abcdefghijklmnopqrstuvwxyz"
	upperCharset  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	symbolCharset = "!@#$%&*"
)

func GenerateCaptcha(length int, captchaType CaptchaType) string {
	var charset string
	if captchaType&CaptchaNumber != 0 {
		charset += numberCharset
	}
	if captchaType&CaptchaUppercase != 0 {
		charset += upperCharset
	}
	if captchaType&CaptchaLowercase != 0 {
		charset += lowerCharset
	}
	if captchaType&CaptchaSymbol != 0 {
		charset += symbolCharset
	}
	if charset == "" {
		charset = numberCharset
	}

	charset = stringx.Shuffle(charset)

	code := make([]byte, length)
	for i := range code {
		code[i] = charset[fastrand.Intn(len(charset))]
	}
	return string(code)
}
