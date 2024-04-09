-- Migration:  create_hits
-- Created at: 2024-01-04 09:16:06
-- ==== UP ====

PRAGMA foreign_keys = ON;

BEGIN;
  CREATE TABLE hits (
		id text PRIMARY KEY,
		path text not null,
		query text,
		title text,
		visitor_id text,
		ip text,
		referer text,
		user_agent text,
		browser text,
		os text,
		response_time int,
		time_on_page int,
		width int,
		height int,
		device_pixel_ratio int,
		is_bot boolean not null default 0 check (is_bot in (0,1)),
		created_at int not null default (unixepoch('now') * 1000)
	);
  CREATE INDEX index_hits_path on hits(path);
  CREATE INDEX index_hits_created_at on hits(created_at);
  CREATE INDEX index_hits_visitor_id on hits(visitor_id);
  CREATE INDEX index_hits_ip on hits(ip);
  CREATE INDEX index_hits_browser on hits(browser);
  CREATE INDEX index_hits_title on hits(title);
  CREATE INDEX index_hits_referer on hits(referer);
  CREATE INDEX index_hits_response_time on hits(response_time);
  CREATE INDEX index_hits_time_on_page on hits(time_on_page);
COMMIT;

-- ==== DOWN ====

PRAGMA foreign_keys = OFF;

BEGIN;
  drop table hits;
COMMIT;
