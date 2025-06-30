package volc

import "github.com/volcengine/ve-tos-golang-sdk/v2/tos"

type Tos struct {
	*tos.ClientV2
}

type TosOpts struct {
	accessKeyId     string
	accessKeySecret string
	endpoint        string
	region          string
}

func NewTos(opts *TosOpts) *Tos {
	credential := tos.NewStaticCredentials(opts.accessKeyId, opts.accessKeySecret)
	client, err := tos.NewClientV2(opts.endpoint, tos.WithCredentials(credential), tos.WithRegion(opts.region))
	if err != nil {
		panic(err)
	}
	return &Tos{client}
}
