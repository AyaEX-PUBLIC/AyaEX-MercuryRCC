#!/usr/bin/env bash
# Build and start the RCC proxy and the Orbiter in the background for the
# Mercury Core + Aya deployment. Run from a desktop session: Aya.Server's
# thumbnail renders need the X display (DISPLAY) for their OpenGL context.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
mkdir -p "$ROOT/logs"

(cd "$ROOT/Proxy" && go build -o Proxy .)
(cd "$ROOT/Orbiter" && go build -o Orbiter .)

start() {
	local name="$1"
	if [[ -f "$ROOT/logs/$name.pid" ]] && kill -0 "$(cat "$ROOT/logs/$name.pid")" 2>/dev/null; then
		echo "$name is already running (pid $(cat "$ROOT/logs/$name.pid"))"
		return
	fi
	# exec keeps the background subshell's PID, so $! is the service itself
	(cd "$ROOT/$name" && exec setsid "./$name" >>"$ROOT/logs/$name.log" 2>&1 </dev/null) &
	echo $! >"$ROOT/logs/$name.pid"
	echo "started $name (pid $(cat "$ROOT/logs/$name.pid")), log: logs/$name.log"
}

start Proxy
start Orbiter
