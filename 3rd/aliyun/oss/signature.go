// Package oss 阿里云oss sdk
package oss

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

const (
	ossSignatureVersion = "OSS4-HMAC-SHA256"
	ossServiceName      = "oss"
	signatureVersion    = "aliyun_v4"
	ossRequestVersion   = "aliyun_v4_request"
	sigVerKey           = "x-oss-signature-version"
	credentialKey       = "x-oss-credential"
	dateKey             = "x-oss-date"
	securityTokeKey     = "x-oss-security-token"
	bucketKey           = "bucket"
)

func (o *Oss) getCurrentDateTime(expire int64) (short, long, expired string) {
	now := time.Now()
	short = now.UTC().Format("20060102")
	long = now.UTC().Format("20060102T150405Z")
	expired = time.Unix(now.Unix()+expire, 0).Format("2006-01-02T15:04:05Z")
	return
}

func (o *Oss) calculateHMACSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// 签名步骤：https://help.aliyun.com/zh/oss/developer-reference/signature-version-4-recommend?spm=a2c4g.11186623.0.0.4ed44c47q7mV24
func (o *Oss) signatureV4(policy, accessSecretKeyId, xDate string) (signature string) {
	dateKey := o.calculateHMACSHA256([]byte(signatureVersion+accessSecretKeyId), xDate)
	dateRegionKey := o.calculateHMACSHA256(dateKey, o.regionId)
	dateRegionServiceKey := o.calculateHMACSHA256(dateRegionKey, "oss")
	signingKey := o.calculateHMACSHA256(dateRegionServiceKey, ossRequestVersion)
	sigByte := o.calculateHMACSHA256(signingKey, policy)
	signature = hex.EncodeToString(sigByte)
	return
}

func (o *Oss) getPolicyToken(req *GetPostPolicyReq) (resp *GetPolicyTokenResp, err error) {
	short, long, tokenExpired := o.getCurrentDateTime(req.ExpiredTime)
	ossCredentials := fmt.Sprintf("%s/%s/%s/%s/%s", o.accessKeyId, short, o.regionId, ossServiceName, ossRequestVersion)
	policy := &PostPolicy{
		Expiration: tokenExpired,
		Conditions: []interface{}{
			map[string]string{sigVerKey: ossSignatureVersion},
			map[string]string{credentialKey: ossCredentials},
			map[string]string{dateKey: long},
		},
	}
	if o.securityToken != nil {
		policy.Conditions = append(policy.Conditions, map[string]interface{}{
			securityTokeKey: *o.securityToken,
		})
	}
	if req.Bucket != nil {
		policy.Conditions = append(policy.Conditions, map[string]interface{}{
			bucketKey: *req.Bucket,
		})
	}
	if len(req.ContentTypes) > 0 {
		policy.Conditions = append(policy.Conditions, []interface{}{"in", "$content-type", req.ContentTypes})
	}
	if req.UploadDir != nil {
		policy.Conditions = append(policy.Conditions, []interface{}{"starts-with", "$key", *req.UploadDir})
	}
	if req.MinLength != nil && req.MaxLength != nil {
		policy.Conditions = append(policy.Conditions, []interface{}{"content-length-range", req.MinLength, req.MaxLength})
	}
	if len(req.Conditions) > 0 {
		policy.Conditions = append(policy.Conditions, req.Conditions...)
	}
	resp = &GetPolicyTokenResp{}
	if req.Callback != nil {
		callback, err := NewCallback(req.Callback)
		if err != nil {
			return nil, err
		}
		policy.Conditions = append(policy.Conditions, map[string]string{"callback": callback})
		resp.Callback = callback
	}
	policyJson, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}
	policyBase64 := base64.StdEncoding.EncodeToString(policyJson)
	signature := o.signatureV4(policyBase64, o.accessKeySecret, short)
	resp.Policy = policyBase64
	resp.SignatureVersion = ossSignatureVersion
	resp.Credential = ossCredentials
	resp.Date = long
	resp.Signature = signature
	resp.ExpiredTime = tokenExpired
	if o.securityToken != nil {
		resp.SecurityToken = *o.securityToken
	}
	return
}
