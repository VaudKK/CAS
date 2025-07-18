DO $$
DECLARE
    new_table_name text;
BEGIN
    new_table_name := 'fund_categories_dropped_' || to_char(CURRENT_TIMESTAMP, 'YYYYMMDD_HH24MI_SS');
    EXECUTE 'ALTER TABLE fund_categories RENAME TO ' || new_table_name;
END $$;