# MercuryRCC (Aya fork): agent instructions

Mercury's gameserver services, adapted to run native Linux **Aya.Server**
instead of RCCService.exe under Wine:

- `Orbiter/`: starts and stops gameservers for the Site (private API :64991)
  and serves the public status stream and the WebSocket proxy to running games
  (:64992, `/play/{id}`).
- `Proxy/`: runs Aya.Server processes and sends them RCC SOAP jobs: hosted games
  (OpenJobEx, RenewLease, CloseJob) and thumbnail renders (BatchJobEx) (:64990).

## Always commit and push

- **Commit and push every change to this repository as soon as it's done and
  verified**: one commit per addition or fix, with a message saying what changed
  and why. Don't leave work uncommitted or unpushed.
- Remote `origin` is `http://192.168.1.44:3006/Mercury-AyaWASM/MercuryRCC.git`
  (branch `main`; the local branch tracks it, so `git push` works). `upstream`
  is the original `tp-link-extender/RCCService`.
- The companion repo `../MercuryCore` (the Site) follows the same rule.

## Never commit secrets

`Proxy/.env` and `Orbiter/.env` hold `RCC_KEY`, `PROXY_KEY`, `GAMESERVER_KEY`
and the Aya.Server `--httpAccessKey`. They're git-ignored; keep it that way, and
check diffs for keys before committing. Logs (`logs/`) are ignored too.

The Orbiter sends `GAMESERVER_KEY` as a bearer token when it fetches a place's
host script from the Site (`/game/{id}/serve`, which carries the place's server
ticket), so it must match the one in the Site's `Site/.env`.

## Working on it

- `./run.sh` builds both services and starts them in the background (logs in
  `logs/`); `./stop.sh` stops them and every Aya.Server they started. Run them
  from the desktop session: renders need `DISPLAY`.
- Hosted gameservers run in a hardened container (`Gameserver/Dockerfile`, on
  Docker Hardened Images Debian 13) when `GAMESERVER_IMAGE` is set in
  `Proxy/.env`; renders stay native. Each game gets a fresh container: a non-root
  user (the one running the proxy), no capabilities, a read-only filesystem, memory/CPU/process
  limits, and only Aya's build and resources (read-only) plus the DataStore
  folder (`data/datastore`, git-ignored) mounted. It shares the host network.
  `Gameserver/build.sh` builds the image (needs `docker login dhi.io` once);
  rebuild it if a new Aya.Server links a new library.
- Restarting disconnects every player. Check `curl localhost:64991/` for running
  gameservers first.
- `go vet ./...` and `gofmt -l .` in `Orbiter/` and `Proxy/` before committing.
- Aya.Server's options come from `RCC_ARGS` in `Proxy/.env`: WebSocket
  transport, content path, the `--httpAllowHost`/`--scriptHttpAllowHost`
  allowlist and `--httpAccessKey`. `MercuryCore/aya/set-domain.sh` updates the
  allowlist's public domain.
- The engine source is the separate `aya-dev` repository
  (`build/x64-linux-rel/bin/Aya.Server`).
