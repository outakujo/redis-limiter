local key = KEYS[1]
local maxRequests = tonumber(ARGV[1])
local timeWindow = tonumber(ARGV[2])
if redis.call("EXISTS", key) == 0 then
    redis.call("SET", key, 1)
    redis.call("EXPIRE", key, timeWindow)
    return 1
else
    if redis.call("INCR", key) > maxRequests then
        return 0
    else
        return 1
    end
end