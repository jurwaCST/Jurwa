#!/bin/bash
# Runs every schema file in its own database inside the pgtest container.
# Each file runs twice to prove it is safe to re-run (IF NOT EXISTS, ON CONFLICT).
# Usage (from the repo root):
#   docker cp docs/design/schema pgtest:/schema
#   docker exec -i pgtest bash /schema/test-schema.sh
exec 2>&1
# Hide harmless "already exists, skipping" notices; real errors are still shown.
export PGOPTIONS="--client-min-messages=warning"
fail=0
for f in /schema/*.sql; do
  name=$(basename "$f" .sql)
  db="${name//-/_}"
  [ "$db" = "authorization" ] && db="authorization_db"
  psql -q -U postgres -c "DROP DATABASE IF EXISTS $db;" >/dev/null
  psql -q -U postgres -c "CREATE DATABASE $db;" >/dev/null
  if psql -q -U postgres -d "$db" -v ON_ERROR_STOP=1 -f "$f" >/dev/null &&
     psql -q -U postgres -d "$db" -v ON_ERROR_STOP=1 -f "$f" >/dev/null; then
    tables=$(psql -U postgres -d "$db" -At -c "SELECT string_agg(table_name, ', ' ORDER BY table_name) FROM information_schema.tables WHERE table_schema = 'public';")
    echo "OK    $name.sql -> database $db: $tables"
  else
    echo "FAIL  $name.sql"
    fail=1
  fi
done
if [ $fail -eq 0 ]; then echo "All schema files ran with no errors."; else echo "Some files failed, see above."; fi
exit $fail
