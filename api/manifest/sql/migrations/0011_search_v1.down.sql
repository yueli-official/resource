DROP TABLE IF EXISTS search_batch_receipts;
DROP TABLE IF EXISTS search_documents;
DROP TABLE IF EXISTS search_generations;
DROP TABLE IF EXISTS search_instances;

DROP TRIGGER IF EXISTS resources_search_revision ON resources;
DROP FUNCTION IF EXISTS resource_bump_search_revision();
ALTER TABLE resources DROP COLUMN IF EXISTS search_revision;
DROP TEXT SEARCH CONFIGURATION IF EXISTS chinese_zh;
DROP EXTENSION IF EXISTS zhparser;
