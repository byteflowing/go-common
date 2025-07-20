package redis

const unLockLua = `
	if
		redis.call("GET", KEYS[1]) == ARGV[1]
	then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
`

const renewLockLua = `
		if
			redis.call("GET", KEYS[1]) == ARGV[1]
		then
			return redis.call("PEXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
`

const incrWithExpireLua = `
	local current = redis.call("INCR", KEYS[1])
	if current == 1 then
		redis.call("EXPIRE", KEYS[1], ARGV[1])
	end
	return current
`

// Lua 返回：
//   - 1 → 允许请求
//   - 100 → 命中 index 0
//   - 101 → 命中 index 1
//   - 102 → 命中 index 2
const slidingWindowLua = `
	-- KEYS[1] = 限流 key
	-- ARGV[1] = 当前时间戳（秒）
	-- ARGV[2] = 唯一 ID（避免重复）
	-- ARGV[3] = TTL（最长窗口）
	-- ARGV[4...n] = 多窗口配置，每两个参数一组：窗口秒数、限制次数

	local now = tonumber(ARGV[1])
	local id = ARGV[2]
	local ttl = tonumber(ARGV[3])
	
	-- 清除所有超过最大窗口的数据
	redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", now - ttl)
	
	-- 遍历所有窗口，判断限流
	for i = 4, #ARGV, 2 do
		local window = tonumber(ARGV[i])
		local limit = tonumber(ARGV[i + 1])
		local count = redis.call("ZCOUNT", KEYS[1], now - window, now)
		if count >= limit then
			-- 返回窗口下标(从100开始)
			local index = (i - 4) / 2
			return index + 100
		end
	end
	
	-- 添加当前请求记录
	-- 多加5buffer防止边界提前删除
	local added = redis.call("ZADD", KEYS[1], now, id)
	if added == 1 then
		redis.call("EXPIRE", KEYS[1], ttl + 5)
	end
	return 1
`
