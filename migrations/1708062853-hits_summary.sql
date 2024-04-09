-- Migration:  hits_summary
-- Created at: 2024-02-15 21:54:13
-- ==== UP ====

PRAGMA foreign_keys = ON;

BEGIN;
  CREATE TABLE hits_summary (
		id int PRIMARY KEY,
		path text not null,
		query text,
		title text,
		response_time int,
		time_on_page int,
		is_bot boolean not null default 0 check (is_bot in (0,1)),
    num int not null,
		date int not null default (unixepoch('now'))
	);
COMMIT;

-- ==== DOWN ====

PRAGMA foreign_keys = OFF;

BEGIN;
  DROP TABLE hits_summary;
COMMIT;
