CREATE OR REPLACE PROCEDURE update_handles_in_batches()
LANGUAGE plpgsql
AS $$
DECLARE
    batch_size INT := 10000;
    start_id BIGINT;
    current_batch_end_id BIGINT;
    max_id BIGINT;
    rows_affected INT;
BEGIN
    -- Use a temporary table for progress tracking to survive disconnects.
    CREATE TEMPORARY TABLE IF NOT EXISTS batch_progress (
        last_processed_id BIGINT NOT NULL
    );

    -- Initialize if empty
    IF NOT EXISTS (SELECT 1 FROM batch_progress) THEN
        INSERT INTO batch_progress (last_processed_id) VALUES (0);
    END IF;

    SELECT last_processed_id INTO start_id FROM batch_progress LIMIT 1;
    SELECT COALESCE(max(id), 0) INTO max_id FROM account_infos;

    WHILE start_id < max_id LOOP
        current_batch_end_id := start_id + batch_size;
        RAISE NOTICE 'Processing batch from id > % to <= %', start_id, current_batch_end_id;

        UPDATE account_infos
        SET handle = substring(jsonb_extract_path_text(describe, 'didDoc', 'alsoKnownAs', '0') from 'at://(.*)')
        WHERE
            id > start_id AND id <= current_batch_end_id
            AND (handle IS NULL OR handle = '')
            AND jsonb_path_exists(describe, '$.didDoc.alsoKnownAs[0]')
            AND substring(jsonb_extract_path_text(describe, 'didDoc', 'alsoKnownAs', '0') from 'at://(.*)') IS NOT NULL;

        GET DIAGNOSTICS rows_affected = ROW_COUNT;
        RAISE NOTICE 'Updated % rows in this batch.', rows_affected;

        -- Update progress and commit to make it durable
        UPDATE batch_progress SET last_processed_id = current_batch_end_id;
        COMMIT;

        start_id := current_batch_end_id;
    END LOOP;

    RAISE NOTICE 'Processing complete.';
END;
$$;

-- To run the procedure:
-- CALL update_handles_in_batches();