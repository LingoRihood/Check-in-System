-- KEYS[1]: 用户签到key
-- ARGV[1]: 当前日偏移量(从年初开始的天数), 即从年初开始的今天是第几天（例如：1 月 1 日为 1，1 月 2 日为 2，以此类推）。
-- ARGV[2]: 提醒阈值, 即最多检查多少天之前的签到状态来决定是否给用户提醒。

-- offset 是当前日期的偏移量（从年初开始的第几天）
local offset = tonumber(ARGV[1])

-- threshold 是提醒阈值（多少天内连续签到才不提醒）
local threshold = tonumber(ARGV[2])

-- 检查今天是否已签到(最新位)
local today = redis.call('GETBIT', KEYS[1], offset)
if today == 1 then return 0 end  -- 已签到无需提醒

-- 检查最近threshold天的签到情况
local continuous = true
for i = 1, threshold do
    local bit = redis.call('GETBIT', KEYS[1], offset - i)
    -- 如果某一天的签到状态不是 1（即 bit ~= 1），表示用户某天没有签到，将 continuous 设为 false，并退出循环（break）
    if bit ~= 1 then
        continuous = false
        break
    end
end

-- 返回结果：1需要提醒 0不需要
return continuous and 1 or 0
-- 如果 continuous 为 true：continuous and 1 会返回 1，然后 or 0 不会影响，最终返回 1。
-- 如果 continuous 为 false：continuous and 1 会返回 false，然后 or 0 会使整个表达式的结果变成 0