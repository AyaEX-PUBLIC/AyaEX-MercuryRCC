#!/usr/bin/env bash
# Builds the hardened gameserver image (Dockerfile) as GAMESERVER_IMAGE from
# Proxy/.env (mercury-gameserver by default), using the Aya.Server build that
# RCC_EXE points at. Pulling Docker Hardened Images needs `docker login dhi.io`
# once, with a Docker Hub account and a personal access token.
# Extra arguments go to docker build, e.g. --build-arg BASE=... --no-cache
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# a setting from Proxy/.env, or nothing if it isn't there
setting() { { grep -E "^$1=" "$ROOT/Proxy/.env" || true; } | tail -1 | cut -d= -f2-; }
exe="$(setting RCC_EXE)"
image="$(setting GAMESERVER_IMAGE)"
image="${image:-mercury-gameserver}"
[[ -x "$exe" ]] || { echo "RCC_EXE in Proxy/.env isn't an Aya.Server build: $exe" >&2; exit 1; }

docker build -t "$image" --build-context aya="$(dirname "$exe")" "$@" "$ROOT/Gameserver"
echo "built $image; restart the RCC services (stop.sh, run.sh) to use it"
