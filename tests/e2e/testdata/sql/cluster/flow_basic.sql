-- FIXME(zyy17): The test case will be replaced by the real sqlness test case.

CREATE TABLE `ngx_access_log` (
  `client` STRING NULL,
  `ua_platform` STRING NULL,
  `referer` STRING NULL,
  `method` STRING NULL,
  `endpoint` STRING NULL,
  `trace_id` STRING NULL FULLTEXT,
  `protocol` STRING NULL,
  `status` SMALLINT UNSIGNED NULL,
  `size` DOUBLE NULL,
  `agent` STRING NULL,
  `access_time` TIMESTAMP(3) NOT NULL,
    TIME INDEX (`access_time`)
    )
WITH(
    append_mode = 'true'
);

CREATE TABLE `ngx_statistics` (
  `status` SMALLINT UNSIGNED NULL,
  `total_logs` BIGINT NULL,
  `min_size` DOUBLE NULL,
  `max_size` DOUBLE NULL,
  `avg_size` DOUBLE NULL,
  `high_size_count` DOUBLE NULL,
  `time_window` TIMESTAMP time index,
  `update_at` TIMESTAMP NULL,
   PRIMARY KEY (`status`));

CREATE FLOW ngx_aggregation
SINK TO ngx_statistics
AS
SELECT
    status,
    count(client) AS total_logs,
    min(size) as min_size,
    max(size) as max_size,
    avg(size) as avg_size,
    sum(case when `size` > 550::double then 1::double else 0::double end) as high_size_count,
    date_bin(INTERVAL '1 minutes', access_time) as time_window,
FROM ngx_access_log
GROUP BY
    status,
    time_window;

INSERT INTO ngx_access_log
VALUES
    ("android", "Android", "referer", "GET", "/api/v1", "trace_id", "HTTP", 200, 1000, "agent", "2021-07-01 00:00:01.000"),
    ("ios", "iOS", "referer", "GET", "/api/v1", "trace_id", "HTTP", 200, 500, "agent", "2021-07-01 00:00:30.500"),
    ("android", "Android", "referer", "GET", "/api/v1", "trace_id", "HTTP", 200, 600, "agent", "2021-07-01 00:01:01.000"),
    ("ios", "iOS", "referer", "GET", "/api/v1", "trace_id", "HTTP", 404, 700, "agent", "2021-07-01 00:01:01.500");

SELECT * FROM ngx_statistics;

INSERT INTO ngx_access_log
VALUES
    ("android", "Android", "referer", "GET", "/api/v1", "trace_id", "HTTP", 200, 500, "agent", "2021-07-01 00:01:01.000"),
    ("ios", "iOS", "referer", "GET", "/api/v1", "trace_id", "HTTP", 404, 800, "agent", "2021-07-01 00:01:01.500");

SELECT * FROM ngx_statistics;
