DO $$
DECLARE
    new_table_name text;
BEGIN
    new_table_name := 'reset_links_dropped_' || to_char(CURRENT_TIMESTAMP, 'YYYYMMDD_HH24MI_SS');
    EXECUTE 'ALTER TABLE reset_links RENAME TO ' || new_table_name;
END $$;