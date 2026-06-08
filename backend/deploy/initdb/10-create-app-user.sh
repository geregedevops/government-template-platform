#!/bin/sh
# Runs ONCE, at first database initialisation, by the postgres image
# entrypoint — executed as the bootstrap superuser POSTGRES_USER against
# POSTGRES_DB (before the migrate container runs).
#
# Why: the users table enforces Row-Level Security (migration 6), but a
# SUPERUSER (which the postgres image makes POSTGRES_USER) BYPASSES RLS even
# when it is FORCED. So we create a NON-superuser role for the API to connect
# as; then the RLS policies actually take effect. The migrate container keeps
# using POSTGRES_USER (it needs superuser for CREATE EXTENSION "uuid-ossp").
#
# FAIL-CLOSED: APP_DB_USER / APP_DB_PASSWORD ЗААВАЛ шаардлагатай. Дутуу бол
# init-ийг зогсооно (exit 1) — өмнө нь чимээгүй superuser-ээр fallback хийж
# бүх RLS-ийг алгасдаг байсан (аюултай).
set -e

if [ -z "${APP_DB_USER}" ] || [ -z "${APP_DB_PASSWORD}" ]; then
  echo "initdb: FATAL — APP_DB_USER/APP_DB_PASSWORD must be set (refusing to run the API as a superuser that bypasses RLS)." >&2
  exit 1
fi

# Нууц үгийг SQL-д шууд залгахгүй (injection эрсдэл) — psql -v хувьсагчаар
# дамжуулж, кодод параметржүүлж хэрэглэнэ.
psql -v ON_ERROR_STOP=1 \
  --username "${POSTGRES_USER}" --dbname "${POSTGRES_DB}" \
  -v app_user="${APP_DB_USER}" -v app_pass="${APP_DB_PASSWORD}" <<SQL
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = :'app_user') THEN
    EXECUTE format('CREATE ROLE %I LOGIN PASSWORD %L NOSUPERUSER NOBYPASSRLS', :'app_user', :'app_pass');
  END IF;
END
\$\$;

GRANT USAGE ON SCHEMA public TO :"app_user";

-- Tables are created later by the migrate container (running as
-- ${POSTGRES_USER}); grant DML on those future tables automatically so the
-- API role never needs superuser. RLS still constrains which ROWS it sees.
ALTER DEFAULT PRIVILEGES FOR ROLE "${POSTGRES_USER}" IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO :"app_user";
ALTER DEFAULT PRIVILEGES FOR ROLE "${POSTGRES_USER}" IN SCHEMA public
  GRANT USAGE, SELECT ON SEQUENCES TO :"app_user";
SQL

echo "initdb: created non-superuser role '${APP_DB_USER}' for the API (RLS enforced)."
