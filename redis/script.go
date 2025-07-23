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
			return redis.call("PEXPIRE", KEYS[1], tonumber(ARGV[2]))
		else
			return 0
		end
`

const incrWithExpireLua = `
	local current = redis.call("INCR", KEYS[1])
	if current == 1 then
		redis.call("PEXPIRE", KEYS[1], tonumber(ARGV[1]))
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
	-- ARGV[2] = 当前请求唯一ID（避免重复）
	-- ARGV[3] = TTL（ZSET 过期时间）
	-- ARGV[4...n] = 多窗口配置，每两个参数一组：窗口秒数、限制次数

	local now = tonumber(ARGV[1])
	local id = ARGV[2]
	local ttl = tonumber(ARGV[3])
	-- 参数校验
	if not now or not ttl then
		return -1
	end
	-- 清除所有超过最大窗口的数据
	redis.call("ZREMRANGEBYSCORE", KEYS[1], "-inf", now - ttl)
	
	-- 遍历限流窗口配置，每组为：窗口大小、最大请求数
	for i = 4, #ARGV, 2 do
		local window = tonumber(ARGV[i])
		local limit = tonumber(ARGV[i + 1])
		local count = redis.call("ZCOUNT", KEYS[1], now - window, now)
		if count >= limit then
			-- 返回窗口下标(从100开始)
			-- 返回100、101、102等
			return 100 + (i - 4) / 2
		end
	end
	
	-- 添加当前请求记录
	redis.call("ZADD", KEYS[1], now, id)
	-- 每次请求重置TTL
	redis.call("EXPIRE", KEYS[1], ttl)
	return 1
`

const allowFixedLimitLua = `
	-- KEYS[1] = 限流 key（如 user:123:api:send_code）
	-- ARGV[1] = 限流周期（秒）
	-- ARGV[2] = 限流次数上限
	local ttl = tonumber(ARGV[1])
	local limit = tonumber(ARGV[2])
	local current = redis.call("INCR", KEYS[1])
	if current == 1 then
		redis.call("PEXPIRE", KEYS[1], ttl)
	end
	if current > limit then
		return 0
	end
	return 1
`

const allowDailyLimitLua = `
	-- KEYS[1]: 限流 key，比如 limit:api:xxx:20250721
	-- ARGV[1]: 今日剩余毫秒数
	-- ARGV[2]: 限流最大次数
	
	local current = redis.call("INCR", KEYS[1])
	if current == 1 then
		redis.call("PEXPIRE", KEYS[1], tonumber(ARGV[1]))
	end
	if current > tonumber(ARGV[2]) then
		return 0
	end
	return 1
`
