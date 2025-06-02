---@module 'mygrep.alias.test'
---@brief LEER
---@desc LEER
---@class MygrepTestDef
---@field get_session_requests fun(): { successful: number, failed: number, cache_hitted: number, fcache_hitted: number } Retrieves the current session request counts
---@field get_total_requests fun(): { successful: number, failed: number, cache_hitted: number, fcache_hitted: number } Retrieves the total request counts from the file
---@field increase_success fun(uuid: any, query: any, source: any, context: any, duration_ms: any, status_code: any): any Increases the successful request count
---@field increase_failed fun(uuid: any, query: any, source: any, context: any, duration_ms: any, status_code: any, error: any): any Increases the failed request count
---@field increase_cache_hit fun(uuid: any, query: any, source: any, context: any): any Increases the cache hit count
---@field increase_fcache_hit fun(uuid: any, query: any, source: any, context: any): any Increases the cache hit count
---@field check_rate_limit fun(): any Checks the current GitHub rate limit and displays a warning if low
---@field record_metrics fun(): boolean Returns the current record metrics state
---@field toggle_record_metrics fun(): boolean Toogle the current record metrics state
---@field set_record_metrics fun(bool: boolean): boolean Set the current record metrics state
---@field req_count ReqCount
---@field rate_limits RateLimitDetails

---@class MygrepTest : MygrepTestDef
local M = {}
