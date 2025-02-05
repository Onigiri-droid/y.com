#!/usr/bin/env bash
# Ожидает, пока указанный хост и порт станут доступными
host="$1"
shift
port="$1"
shift

echo "Ожидание доступности $host:$port..."
while ! nc -z "$host" "$port"; do
  sleep 1
done

echo "$host:$port доступен. Запуск приложения..."
exec "$@"
