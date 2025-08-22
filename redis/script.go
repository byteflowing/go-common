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

const slidingWindowLua = `
for i = 1, #KEYS do
	local key = KEYS[i]
	local duration = tonumber(ARGV[(i - 1) * 2 + 1])
	local limit = tonumber(ARGV[(i - 1) * 2 + 2])
	local cnt = redis.call("INCR", key)
	if cnt == 1 then
		redis.call("EXPIRE", key, duration)
	end
	if cnt > limit then
		local ttl = redis.call("TTL", key)
		if ttl >= 0 then
			return {100 + (i - 1), ttl}
		end
	end
end
return {1, 0}
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
