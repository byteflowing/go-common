package idx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCaptcha(t *testing.T) {
	length := 8
	number := GenerateCaptcha(length, CaptchaNumber)
	letter := GenerateCaptcha(length, CaptchaLetter)
	numberLetter := GenerateCaptcha(length, CaptchaNumberLetter)
	letterUppercase := GenerateCaptcha(length, CaptchaLetterUppercase)
	letterLowercase := GenerateCaptcha(length, CaptchaLetterLowercase)
	numberLetterUppercase := GenerateCaptcha(length, CaptchaNumberLetterUppercase)
	numberLetterLowercase := GenerateCaptcha(length, CaptchaNumberLetterLowercase)

	t.Logf(number)
	t.Logf(letter)
	t.Logf(numberLetter)
	t.Logf(letterUppercase)
	t.Logf(letterLowercase)
	t.Logf(numberLetterUppercase)
	t.Logf(numberLetterLowercase)
	assert.Equal(t, length, len(number))
	assert.Equal(t, length, len(letter))
	assert.Equal(t, length, len(numberLetter))
	assert.Equal(t, length, len(letterUppercase))
	assert.Equal(t, length, len(letterLowercase))
	assert.Equal(t, length, len(numberLetterUppercase))
	assert.Equal(t, length, len(numberLetterLowercase))
}
