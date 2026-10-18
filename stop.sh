#!/usr/bin/env bash
# Stop the Orbiter, the RCC proxy and the Aya.Server instances the proxy started.
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
for name in Orbiter Proxy; do
	pidfile="$ROOT/logs/$name.pid"
	if [[ -f "$pidfile" ]]; then
		kill "$(cat "$pidfile")" 2>/dev/null && echo "stopped $name"
		rm -f "$pidfile"
	fi
done
# gameservers running in containers (GAMESERVER_IMAGE in Proxy/.env)
containers="$(docker ps --quiet --filter label=mercury.gameserver 2>/dev/null || true)"
if [[ -n "$containers" ]]; then
	docker stop --time 5 $containers >/dev/null && echo "stopped gameserver containers"
fi
# Aya.Server instances started by the proxy are run with --port <SOAP port> first
if pkill -f 'Aya\.Server --port '; then
	# Aya.Server can ignore SIGTERM once its render thread is up
	sleep 2
	pkill -9 -f 'Aya\.Server --port ' || true
	echo "stopped Aya.Server instances"
fi
