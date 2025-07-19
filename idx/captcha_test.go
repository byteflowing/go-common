package idx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCaptcha_NumberOnly(t *testing.T) {
	code := GenerateCaptcha(6, CaptchaNumber)
	assert.Len(t, code, 6)
	assert.Regexp(t, `^[0-9]{6}$`, code)
}

func TestGenerateCaptcha_UppercaseOnly(t *testing.T) {
	code := GenerateCaptcha(8, CaptchaUppercase)
	assert.Len(t, code, 8)
	assert.Regexp(t, `^[A-Z]{8}$`, code)
}

func TestGenerateCaptcha_LowercaseOnly(t *testing.T) {
	code := GenerateCaptcha(5, CaptchaLowercase)
	assert.Len(t, code, 5)
	assert.Regexp(t, `^[a-z]{5}$`, code)
}

func TestGenerateCaptcha_NumberUppercase(t *testing.T) {
	code := GenerateCaptcha(10, CaptchaNumber|CaptchaUppercase)
	assert.Len(t, code, 10)
	assert.Regexp(t, `^[0-9A-Z]{10}$`, code)
}

func TestGenerateCaptcha_AllTypes(t *testing.T) {
	code := GenerateCaptcha(12, CaptchaNumber|CaptchaUppercase|CaptchaLowercase)
	assert.Len(t, code, 12)
	assert.Regexp(t, `^[0-9a-zA-Z]{12}$`, code)
}

func TestGenerateCaptcha_SymbolIncluded(t *testing.T) {
	code := GenerateCaptcha(8, CaptchaNumber|CaptchaSymbol)
	assert.Len(t, code, 8)
	assert.Regexp(t, `^[0-9!@#\$%&\*]{8}$`, code)
}

func TestGenerateCaptcha_DefaultFallback(t *testing.T) {
	code := GenerateCaptcha(6, 0)
	assert.Len(t, code, 6)
	assert.Regexp(t, `^[0-9]{6}$`, code)
}
