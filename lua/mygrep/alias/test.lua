local M = {}


-- Project-specific Configuration and Utility Modules
local config = require("reposcope.config")
local notify = require("reposcope.utils.debug").notify


---@class ReqCount Counts API requests for profiling purposes
---@field successful number Stores the count of successful API requests for the current session
---@field failed number Stores the count of failed API requests for the current session
---@field cache_hitted number Stores the count of cache hits for the current session
---@field fcache_hitted number Stores the count of filecache hits for the current session
M.req_count = {
  successful = 0,    -- Successful API requests in this session
  failed = 0,        -- Failed API requests in this session
  cache_hitted = 0,  -- Cache hits in this session
  fcache_hitted = 0, -- Filecache hits in this session
}

---@class RateLimits
---@field core RateLimitDetails The rate limit details for the GitHub Core API (general API requests)
---@field search RateLimitDetails The rate limit details for the GitHub Search API (search-related API requests)
---@class RateLimitDetails
---@field limit number The maximum number of requests allowed
---@field remaining number The remaining requests in the current rate limit window
---@field reset number The UNIX timestamp when the rate limit will reset
M.rate_limits = {
  core = {
    limit = 0,     -- The maximum number of requests allowed for the Core API
    remaining = 0, -- The remaining requests available in this session
    reset = 0      -- The timestamp for rate limit reset (UNIX time)
  },
  search = {
    limit = 0,     -- The maximum number of requests allowed for the Search API
    remaining = 0, -- The remaining requests available in this session
    reset = 0      -- The timestamp for rate limit reset (UNIX time)
  }
}

--- Retrieves the current session request counts
---@return { successful: number, failed: number, cache_hitted: number, fcache_hitted: number }
function M.get_session_requests()
  return {
    successful = M.req_count.successful,
    failed = M.req_count.failed,
    cache_hitted = M.req_count.cache_hitted,
    fcache_hitted = M.req_count.fcache_hitted
  }
end

--- Retrieves the total request counts from the file
---@return { successful: number, failed: number, cache_hitted: number, fcache_hitted: number }
function M.get_total_requests()
  local log_path = config.get_option("logfile_path")
  if not log_path then
    notify("[reposcope] Stats not available, logfile path invalid", 4)
    return { successful = 0, failed = 0, cache_hitted = 0, fcache_hitted = 0 }
  end

  if not vim.fn.filereadable(log_path) then
    return { successful = 0, failed = 0, cache_hitted = 0, fcache_hitted = 0 }
  end

  ---@diagnostic disable-next-line vim.loop or vim.us fs_stat exists
  local file_stats = vim.loop.fs_stat(log_path)
  if file_stats and file_stats.size == 0 then
    return { successful = 0, failed = 0, cache_hitted = 0, fcache_hitted = 0 }
  end

  local ok, raw = pcall(vim.fn.readfile, log_path)
  if not ok then
    notify("[reposcope] Error reading stats file: " .. raw, 4)
    return { successful = 0, failed = 0, cache_hitted = 0, fcache_hitted = 0 }
  end
  if #raw == 0 then
    return { successful = 0, failed = 0, cache_hitted = 0, fcache_hitted = 0 }
  end

  local json_data = vim.json.decode(table.concat(raw, "\n")) or {}

  local successful, failed, cache_hitted, fcache_hitted = 0, 0, 0, 0

  for _, log in pairs(json_data) do
    if log.type == "api_success" then
      successful = successful + 1
    elseif log.type == "api_failed" then
      failed = failed + 1
    elseif log.type == "cache_hit" then
      cache_hitted = cache_hitted + 1
    elseif log.type == "filecache_hit" then
      fcache_hitted = fcache_hitted + 1
    end
  end

  return {
    successful = successful,
    failed = failed,
    cache_hitted = cache_hitted,
    fcache_hitted = fcache_hitted
  }
end

--- Logs request details to request_log.json in JSON object format
---@param uuid string request identifier
---@param data table The request data to log
local function log_request(uuid, data)
  -- Sicherstellen, dass uuid ein String ist
  if type(uuid) ~= "string" then
    vim.notify("[reposcope] Invalid UUID type. Expected string, got " .. type(uuid), vim.log.levels.ERROR)
    return
  end

  local log_max = config.options.log_max or 1000
  local log_path = config.get_option("logfile_path")

  if not log_path then
    notify("[reposcope] log_path for log_request() is invalid.", 2)
    return
  end

  vim.schedule(function()
    local logs = {}

    -- Read existing log file if available
    if vim.fn.filereadable(log_path) == 1 then
      local raw = vim.fn.readfile(log_path)      -- string[]
      if raw and not vim.tbl_isempty(raw) then
        local raw_json = table.concat(raw, "\n") -- string
        local success, decoded_logs = pcall(function()
          return vim.json.decode(raw_json)
        end)

        if success and type(decoded_logs) == "table" then
          logs = decoded_logs
        else
          vim.notify("[reposcope] Invalid JSON format in log file. Starting fresh log.", vim.log.levels.WARN)
        end
      end
    end

    -- Ensure logs is always a table
    if type(logs) ~= "table" then
      logs = {}
    end

    -- Add new log entry
    local log_key = uuid .. ":" .. (data.type or "unknown")
    logs[log_key] = data

    -- Remove oldest entry if too many logs exist
    if vim.tbl_count(logs) > log_max then
      local oldest_key = next(logs)
      if oldest_key then
        logs[oldest_key] = nil
      end
    end

    -- Encode and save logs to file with formatted JSON
    local ok, json_or_err = pcall(vim.json.encode, logs)
    if not ok then
      notify("[reposcope] Failed to encode logs to JSON: " .. tostring(json_or_err), 5)
      return
    end
    vim.fn.writefile(vim.split(json_or_err, "\n"), log_path)
  end)
end

--- Increases the successful request count
function M.increase_success(uuid, query, source, context, duration_ms, status_code)
  M.req_count.successful = M.req_count.successful + 1
  log_request(uuid, {
    timestamp = os.date("!%Y-%m-%dT%H:%M:%SZ"),
    type = "api_success",
    query = query,
    source = source,
    context = context,
    duration_ms = duration_ms,
    status_code = status_code
  })
end

--- Increases the failed request count
function M.increase_failed(uuid, query, source, context, duration_ms, status_code, error)
  M.req_count.failed = M.req_count.failed + 1
  log_request(uuid, {
    timestamp = os.date("!%Y-%m-%dT%H:%M:%SZ"),
    type = "api_failed",
    query = query,
    source = source,
    context = context,
    duration_ms = duration_ms,
    status_code = status_code,
    error_message = error
  })
end

--- Increases the cache hit count
function M.increase_cache_hit(uuid, query, source, context)
  M.req_count.cache_hitted = M.req_count.cache_hitted + 1
  log_request(uuid, {
    timestamp = os.date("!%Y-%m-%dT%H:%M:%SZ"),
    type = "cache_hit",
    query = query,
    source = source,
    context = context,
  })
end

--- Increases the cache hit count
function M.increase_fcache_hit(uuid, query, source, context)
  M.req_count.fcache_hitted = M.req_count.fcache_hitted + 1
  log_request(uuid, {
    timestamp = os.date("!%Y-%m-%dT%H:%M:%SZ"),
    type = "filecache_hit",
    query = query,
    source = source,
    context = context,
  })
end

--- Checks the current GitHub rate limit and displays a warning if low
function M.check_rate_limit()
  if M.rate_limits.core.limit > 0 and M.rate_limits.search.limit > 0 then
    local core_used = M.req_count.successful + M.req_count.failed
    local core_remaining = M.rate_limits.core.remaining
    local core_usage = 1 - (core_remaining / M.rate_limits.core.limit)

    if core_usage >= 0.9 then
      vim.schedule(function()
        vim.notify(string.format(
          "[Reposcope] WARNING: GitHub API Core limit critical (%d/%d, remaining: %d)",
          core_used, M.rate_limits.core.limit, core_remaining
        ), 3)
      end)
    elseif core_usage >= 0.75 then
      vim.schedule(function()
        vim.notify(string.format(
          "[Reposcope] INFO: GitHub API Core limit approaching (%d/%d, remaining: %d)",
          core_used, M.rate_limits.core.limit, core_remaining
        ), 2)
      end)
    end

    local search_remaining = M.rate_limits.search.remaining
    local search_usage = 1 - (search_remaining / M.rate_limits.search.limit)

    if search_usage >= 0.9 then
      vim.schedule(function()
        vim.notify(string.format(
          "[Reposcope] WARNING: GitHub API Search limit critical (remaining: %d)",
          search_remaining
        ), 3)
      end)
    elseif search_usage >= 0.75 then
      vim.schedule(function()
        vim.notify(string.format(
          "[Reposcope] INFO: GitHub API Search limit approaching (remaining: %d)",
          search_remaining
        ), 2)
      end)
    end

    return
  end

  -- If no limits set request api and cache it in ram
  local http = require("reposcope.network.http")
  local token = config.options.github_token

  local headers = { "Accept: application/vnd.github+json" }
  if token then
    table.insert(headers, "Authorization: Bearer " .. token)
  end

  http.get("https://api.github.com/rate_limit", function(response)
    if not response then
      vim.schedule(function()
        vim.notify("[Reposcope] Failed to fetch GitHub rate limit.", 4)
      end)
      return
    end

    local data = vim.json.decode(response)
    if data and data.resources then
      M.rate_limits.core.limit = data.resources.core.limit
      M.rate_limits.core.remaining = data.resources.core.remaining
      M.rate_limits.core.reset = data.resources.core.reset

      M.rate_limits.search.limit = data.resources.search.limit
      M.rate_limits.search.remaining = data.resources.search.remaining
      M.rate_limits.search.reset = data.resources.search.reset
    end
  end)
end

---Returns the current record metrics state
---@return boolean The current record metrics state
function M.record_metrics()
  return config.options.metrics
end

---Toogle the current record metrics state
---@return boolean The current record metrics state
function M.toggle_record_metrics()
  config.options.metrics = not config.options.metrics
  return config.options.metrics
end

---Set the current record metrics state
---@param bool boolean Boolean value to set record metrics state
---@return boolean The current record metrics state
function M.set_record_metrics(bool)
  config.options.metrics = bool
  return config.options.metrics
end

return M
