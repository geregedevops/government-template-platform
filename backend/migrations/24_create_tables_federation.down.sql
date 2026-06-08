DELETE FROM permissions WHERE key = 'fed.manage';
DROP TABLE IF EXISTS fed_inbox;
DROP TABLE IF EXISTS fed_outbox;
DROP TABLE IF EXISTS fed_peers;
