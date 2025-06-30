package lbs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/byteflowing/go-common/jsonx"
)

type Opts struct {
	Key string
}

type Map struct {
	key        string
	httpClient *http.Client
}

func NewMap(opts *Opts) *Map {
	return &Map{
		key:        opts.Key,
		httpClient: http.DefaultClient,
	}
}

// PlaceSearch
// 提供多种搜索功能：
//
//	指定城市/区域搜索：如在北京搜索景点。新增高级参数：支持获取车站、机场、园区等较大范围地点的子点和出入口热度，辅助用户选择准确目的地。
//	周边搜索：如，搜索颐和园附近半径500米内的酒店（一个圆形范围）；
//	矩形范围搜索：在地图应用中，往往用于视野内搜索，因为显示地图的区域是个矩形。
//	多边形范围搜索：自定义多边形范围进行地点搜索，由此您可以更好地控制搜索范围的准确性。
//	周边推荐：只需提供中心点及半径（无须关键词），即可搜索获取周边高热度地点，一般用于发送位置、地点签到等场景，自动为用户提供备选地点列表
//	POI详情：通过POI ID查询POI信息
//
// 文档：https://lbs.qq.com/service/webService/webServiceGuide/search/webServiceSearch
func (m *Map) PlaceSearch(ctx context.Context, req *PlaceSearchReq) (resp *PlaceSearchResp, err error) {
	params := url.Values{}
	params.Add("keyword", req.Keyword)
	if req.Boundary != nil {
		boundary := fmt.Sprintf("nearby(%v,%v,%v,%v)", req.Boundary.Latitude, req.Boundary.Longitude, req.Boundary.Radius, req.Boundary.AutoExtend)
		params.Add("boundary", boundary)
	}
	if req.GetSubpois != nil {
		v := "0"
		if *req.GetSubpois {
			v = "1"
		}
		params.Add("get_subpois", v)
	}
	if req.Filter != nil {
		if len(req.Filter.Category) > 0 {
			category := strings.Join(req.Filter.Category, ",")
			v := fmt.Sprintf("category=%s", category)
			params.Add("filter", v)
		}
		if len(req.Filter.Exclude) > 0 {
			category := strings.Join(req.Filter.Exclude, ",")
			v := fmt.Sprintf("category<>%s", category)
			params.Add("filter", v)
		}
	}
	if req.AddedFields != nil {
		params.Add("added_fields", *req.AddedFields)
	}
	if req.OrderBy != nil {
		params.Add("orderby", *req.OrderBy)
	}
	if req.PageSize != nil {
		params.Add("page_size", strconv.Itoa(*req.PageSize))
	}
	if req.PageIndex != nil {
		params.Add("page_index", strconv.Itoa(*req.PageIndex))
	}
	if req.Output != nil {
		params.Add("output", *req.Output)
	}
	if req.Callback != nil {
		params.Add("callback", *req.Callback)
	}
	addr := fmt.Sprintf("%s?%s&key=%s", apiPlaceSearch, params.Encode(), m.key)
	res, err := m.request(addr)
	if err != nil {
		return nil, err
	}
	resp = &PlaceSearchResp{}
	err = jsonx.Unmarshal(res, resp)
	return resp, err
}

// AddressToGeo .
// 本接口提供由文字地址到经纬度的转换能力，并同时提供结构化的省市区地址信息
// 文档：https://lbs.qq.com/service/webService/webServiceGuide/address/Geocoder
func (m *Map) AddressToGeo(ctx context.Context, req *AddressToGeoReq) (resp *AddressToGeoResp, err error) {
	params := url.Values{}
	params.Add("address", req.Address)
	if req.Output != nil {
		params.Add("output", *req.Output)
	}
	if req.Callback != nil {
		params.Add("callback", *req.Callback)
	}
	addr := fmt.Sprintf("%s?%s&key=%s", apiAddrToGeo, params.Encode(), m.key)
	res, err := m.request(addr)
	if err != nil {
		return nil, err
	}
	resp = &AddressToGeoResp{}
	err = jsonx.Unmarshal(res, resp)
	return resp, err
}

func (m *Map) request(reqURL string) ([]byte, error) {
	response, err := m.httpClient.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return io.ReadAll(response.Body)
}
