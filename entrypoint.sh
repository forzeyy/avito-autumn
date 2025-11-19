#!/bin/sh

./wait-for-it.sh db:5432 --timeout=30 --strict -- echo "PostgreSQL поднят"

if /usr/local/bin/sql-migrate up -config=/app/migrate.yaml; then
    echo "миграции применены"
else
    echo "не удалось применить миграции"
    exit 1
fi

echo "запуск приложения"
exec "$@"