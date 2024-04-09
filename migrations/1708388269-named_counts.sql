-- Migration:  generic_counts
-- Created at: 2024-02-19 16:17:49
-- ==== UP ====

PRAGMA foreign_keys = ON;

BEGIN;
  CREATE TABLE named_counts(
    id integer primary key,
    type text not null, -- referer, width x height, browser, os
    name text not null,
    num int not null,
    date int not null default (unixepoch('now'))
);
COMMIT;

-- ==== DOWN ====

PRAGMA foreign_keys = OFF;

BEGIN;
DROP TABLE named_counts;
COMMIT;
