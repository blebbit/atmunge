UPDATE account_infos
SET handle = substring(jsonb_extract_path_text(describe, 'didDoc', 'alsoKnownAs', '0') from 'at://(.*)');