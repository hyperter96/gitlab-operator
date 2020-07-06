PGPASS=$(cat ${POSTGRES_POSTGRES_PASSWORD_FILE}) psql -d gitlabhq_production -U postgres -c 'CREATE EXTENSION pg_trgm;'
