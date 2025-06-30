package lbs

type PlaceSearchReq struct {
	// 搜索关键字，长度最大96个字节
	// 注：keyword仅支持检索一个
	Keyword string `json:"keyword"`
	// 范围
	Boundary *Boundary `json:"boundary"`
	// 是否返回子地点，如大厦停车场、出入口等
	GetSubpois *bool `json:"get_subpois"`
	// 筛选条件
	Filter *Filter `json:"filter"`
	// 返回指定标准附加字段，取值支持：
	// category_code - poi分类编码
	AddedFields *string `json:"added_fields"`
	// 排序，支持按距离由近到远排序，取值：_distance
	// 说明：
	// 1. 周边搜索默认排序会综合考虑距离、权重等多方面因素
	// 2. 设置按距离排序后则仅考虑距离远近，一些低权重的地点可能因距离近排在前面，导致体验下降
	// 取值：_distance
	OrderBy *string `json:"order_by"`
	// 每页条目数，最大限制为20条，默认为10条
	PageSize *int `json:"page_size"`
	// 第x页，默认第1页
	PageIndex *int `json:"page_index"`
	// 返回格式：
	// 支持JSON/JSONP，默认JSON
	Output *string `json:"output"`
	// JSONP方式回调函数
	Callback *string `json:"callback"`
}

type PlaceSearchResp struct {
	Status    int                    `json:"status"`     // 状态码，0为正常，其它为异常
	Message   string                 `json:"message"`    // 状态说明
	Count     int                    `json:"count"`      // 本次搜索结果总数，另外本服务限制最多返回200条数据(data)， 翻页（page_index）页码超过总页数之后返回最后一页的结果
	RequestId string                 `json:"request_id"` // 本次请求的唯一标识，由系统自动生成，用于追查结果有异常时使用
	Data      []*PlaceSearchRespData `json:"data"`       // 搜索结果POI（地点）数组，每项为一个POI（地点）对象
}

type PlaceSearchRespData struct {
	Id           string    `json:"id"`            // POI（地点）唯一标识
	Title        string    `json:"title"`         // POI（地点）名称
	Address      string    `json:"address"`       // 地址
	Tel          string    `json:"tel"`           // 电话
	Category     string    `json:"category"`      // POI（地点）分类
	CategoryCode int64     `json:"category_code"` // POI（地点）分类编码，设置added_fields=category_code时返回
	Type         int       `json:"type"`          // POI类型，值说明：0:普通POI / 1:公交车站 / 2:地铁站 / 3:公交线路 / 4:行政区划
	Location     Location  `json:"location"`      // 坐标
	Distance     float64   `json:"_distance"`     // 距离，单位： 米，在周边搜索、城市范围搜索传入定位点时返回
	AdInfo       *AdInfo   `json:"ad_info"`       // 行政区划信息
	SubPOIS      []*SubPOI `json:"sub_pois"`      // 子地点列表，仅在输入参数get_subpois=1时返回
}

type AddressToGeoReq struct {
	Address  string  `json:"address"`
	Output   *string `json:"output"`
	Callback *string `json:"callback"`
}

type AddressToGeoResp struct {
	Status  int                   `json:"status"`  // 状态码，0为正常，其它为异常
	Message string                `json:"message"` // 状态说明
	Result  *AddressToGeoRespData `json:"result"`
}

type AddressToGeoRespData struct {
	Location          *Location          `json:"location"`           // 解析到的坐标（GCJ02坐标系）
	AddressComponents *AddressComponents `json:"address_components"` // 解析后的地址部件
	AdInfo            *AdInfo            `json:"ad_info"`            // 行政区划信息
	Reliability       int                `json:"reliability"`        // 可信度参考：值范围 1 <低可信> - 10 <高可信>
	Level             int                `json:"level"`              // 解析精度级别，分为11个级别，一般>=9即可采用（定位到点，精度较高） 也可根据实际业务需求自行调整
}

type AddressComponents struct {
	Province     string `json:"province"`      // 省
	City         string `json:"city"`          // 市，如果当前城市为省直辖县级区划，city与district字段均会返回此城市,注：省直辖县级区划adcode第3和第4位分别为9、0，如济源市adcode为419001
	District     string `json:"district"`      // 区，可能为空字串
	Street       string `json:"street"`        // 街道/道路，可能为空字串
	StreetNumber string `json:"street_number"` // 门牌，可能为空字串
}

type Filter struct {
	// 指定分类筛选
	// 分类词数量建议不超过5个，支持设置分类编码
	// e.g. 分类名称：公交车站 分类编码：271013
	Category []string `json:"category"`
	// 排除指定分类
	// 取值同上
	Exclude []string `json:"exclude"`
}

type Boundary struct {
	Latitude   float64 `json:"latitude"`    // 维度
	Longitude  float64 `json:"longitude"`   // 经度
	Radius     float64 `json:"radius"`      // 搜索半径，单位：米，取值范围：10到1000
	AutoExtend bool    `json:"auto_extend"` // [可选] 是否自动扩大范围
}

type Location struct {
	Lat float64 `json:"lat"` // 纬度
	Lng float64 `json:"lng"` // 经度
}

type AdInfo struct {
	// 行政区划代码
	AdCode string `json:"adcode"`
	// 省
	Province string `json:"province"`
	// 市，如果当前城市为省直辖县级区划，此字段会返回为空，由district字段返回
	// 注：省直辖县级区划adcode第3和第4位分别为9、0，如济源市adcode为419001
	City string `json:"city"`
	// 区
	District string `json:"district"`
}

type SubPOI struct {
	ParentId string    `json:"parent_id"` // 主地点ID，对应data中的地点ID
	Id       string    `json:"id"`        // 地点唯一标识
	Title    string    `json:"title"`     // 地点名称
	Tel      string    `json:"tel"`       // 电话
	Category string    `json:"category"`  // POI（地点）分类
	Type     int       `json:"type"`      // POI类型，值说明：0:普通POI / 1:公交车站 / 2:地铁站 / 3:公交线路 / 4:行政区划
	Address  string    `json:"address"`   // 地址
	Location *Location `json:"location"`  // 坐标
	AdInfo   *AdInfo   `json:"ad_info"`   // 行政区划信息
}
