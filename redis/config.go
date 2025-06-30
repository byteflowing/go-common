package redis

type Config struct {
	Addr       string // 连接地址 e.g. 127.0.0.1:6379
	User       string // 用户名
	Password   string // 密码
	DB         int    // db
	Protocol   int    // RESP版本：2或者3
	ClientName string // 客户端名称
}
