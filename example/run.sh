#!/usr/bin/env bash
set -eo pipefail

if [[ -z "${CI}" ]]; then
  trap "exit" INT TERM
  trap "kill 0" EXIT
fi

export GO_GRPC_HMAC_LOG=true
export key_id="key-one"
secret_key="$(head /dev/urandom | LC_ALL=C tr -dc A-Za-z0-9 | head -c24)"
export secret_key

pushd server &>/dev/null
go run . &
sleep 1
popd &>/dev/null

pushd client &>/dev/null
echo "[run.sh] Running client with correct secret"
go run . && echo "[run.sh] Correct secret worked"
export secret_key="wrong-secret"
go run . || echo "[run.sh] Wrong secret failed"
popd &>/dev/null
