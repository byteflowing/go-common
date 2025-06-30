package sts

type CommonResp struct {
	RequestId string
}

type Statement struct {
	Effect    string `json:"Effect"`
	Action    any    `json:"Action"`
	Resource  any    `json:"Resource"`
	Condition any    `json:"Condition"`
}

type Policy struct {
	Version   string       `json:"Version"`
	Statement []*Statement `json:"Statement"`
}

type AssumeRoleReq struct {
	RoleArn         string
	DurationSeconds *int64
	Policy          *Policy
	ExternalId      *string
	RoleSessionName *string
}

type AssumeRoleResp struct {
	Common          *CommonResp
	AccessKeyId     string
	AccessKeySecret string
	SecurityToken   string
	// The time when the STS token expires
	// The time is displayed in UTC
	Expiration string
}
