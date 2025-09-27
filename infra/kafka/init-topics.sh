#!/usr/bin/env bash
set -euo pipefail
BROKERS="${KAFKA_BROKERS:-kafka:9092}"

create () {
  local topic="$1"; shift
  /opt/bitnami/kafka/bin/kafka-topics.sh --create --if-not-exists \
    --bootstrap-server "$BROKERS" --topic "$topic" "$@"
}

# items (normal retention ~ 3d)
create "item" --replication-factor 1 --partitions 1 \
  --config retention.ms=259200000 --config cleanup.policy=delete

# DLQ (uzun retention ~ 14d)
create "item-dlq" --replication-factor 1 --partitions 1 \
  --config retention.ms=1209600000 --config cleanup.policy=delete