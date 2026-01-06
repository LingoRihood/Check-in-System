-- KEYS[1]: 用户签到key (年签到位图，例如 yearSignKeyFormat(userId, year))
-- ARGV[1]: 当前日偏移量（0-based）
--         例如：1月1日为 0，1月2日为 1，以此类推（Go 里传的是 YearDay()-1）
-- ARGV[2]: 提醒阈值 threshold（最多检查 offset 之前连续多少天都签到）

local offset = tonumber(ARGV[1])
local threshold = tonumber(ARGV[2])

-- 参数保护：缺参/非法直接不提醒
if offset == nil or threshold == nil then
    return 0
end
if threshold <= 0 then
    return 0
end
if offset < 0 then
    return 0
end

-- 检查今天是否已签到
local today = redis.call('GETBIT', KEYS[1], offset)
if today == 1 then
    return 0 -- 已签到无需提醒
end

-- ✅ 年初处理：如果历史天数不足 threshold，就按已有天数检查
-- 例如：
-- 1月1日 offset=0 -> checkDays=0（没有昨天）-> 不提醒
-- 1月2日 offset=1, threshold=2 -> checkDays=1（只检查 day0）
local checkDays = threshold
if offset < threshold then
    checkDays = offset
end

-- 如果没有可检查的历史天数（例如 1月1日），不提醒
if checkDays <= 0 then
    return 0
end

-- 检查今天之前 checkDays 天是否连续签到
for i = 1, checkDays do
    local bit = redis.call('GETBIT', KEYS[1], offset - i)
    if bit ~= 1 then
        return 0
    end
end

-- 今天没签，且之前连续（可检查的）天数都签了：需要提醒
return 1
