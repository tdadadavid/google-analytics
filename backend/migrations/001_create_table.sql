-- Postgresql
CREATE TABLE IF NOT EXISTS events (
  id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
  site_id VARCHAR NOT NULL,
  occured_at INT NOT NULL,
  type VARCHAR NOT NULL,
  user_id VARCHAR NOT NULL,
  event VARCHAR NOT NULL,
  category VARCHAR NOT NULL,
  referrer VARCHAR NOT NULL,
  referrer_domain VARCHAR NOT NULL,
  is_touch BOOLEAN NOT NULL,
  browser_name VARCHAR NOT NULL,
  os_name VARCHAR NOT NULL,
  device_type VARCHAR NOT NULL,
  country VARCHAR NOT NULL,
  region VARCHAR NOT NULL,
  timestamp timestamptz DEFAULT now()
);

-- Clickhouse
CREATE TABLE IF NOT EXISTS events (
  site_id String NOT NULL,
  occured_at UInt32 NOT NULL,
  type String NOT NULL,
  user_id String NOT NULL,
  event String NOT NULL,
  category String NOT NULL,
  referrer String NOT NULL,
  referrer_domain String NOT NULL,
  is_touch BOOLEAN NOT NULL,
  browser_name String NOT NULL,
  os_name String NOT NULL,
  device_type String NOT NULL,
  country String NOT NULL,
  region String NOT NULL,
  timestamp Datetime DEFAULT now()
)
ENGINE MergeTree
ORDER BY (site_id, occured_at);