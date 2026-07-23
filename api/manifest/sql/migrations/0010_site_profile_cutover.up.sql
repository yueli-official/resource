-- Resource public Site Profile is moved by the application in the same
-- transaction that narrows resource_site_settings to the resource runtime
-- section. This marker owns the reversible downgrade projection.
SELECT 1;
