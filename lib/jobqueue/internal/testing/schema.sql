-- Copyright (c) https://github.com/maragudk/goqite
-- https://github.com/maragudk/goqite/blob/6d1bf3c0bcab5a683e0bc7a82a4c76ceac1bbe3f/LICENSE
--
-- This source code is licensed under the MIT license found in the LICENSE file
-- in the root directory of this source tree, or at:
-- https://opensource.org/licenses/MIT

create table jobqueue (
  id text primary key default ('m_' || lower(hex(randomblob(16)))),
  created text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  updated text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  queue text not null,
  body blob not null,
  timeout text not null default (strftime('%Y-%m-%dT%H:%M:%fZ')),
  received integer not null default 0
) strict;

create trigger jobqueue_updated_timestamp after update on jobqueue begin
  update jobqueue set updated = strftime('%Y-%m-%dT%H:%M:%fZ') where id = old.id;
end;

create index jobqueue_queue_created_idx on jobqueue (queue, created);