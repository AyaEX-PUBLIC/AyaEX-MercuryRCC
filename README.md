# AyaEX MercuryRCC

The gameserver services of [AyaEX](https://github.com/AyaEX-PUBLIC/AyaEX-MercuryCore),
a fork of Mercury's [RCCService](https://github.com/tp-link-extender/RCCService)
that runs the native Linux AyaEX `Aya.Server` instead of RCCService.exe under
Wine:

- **Orbiter** (`Orbiter/`) starts and stops gameservers for the website. Its
  private API is on :64991. On :64992 it serves the public status stream and
  the WebSocket proxy that browser Players join games through (`/play/{id}`).
- **RCC proxy** (`Proxy/`) runs the `Aya.Server` processes on :64990 and sends
  them RCC SOAP jobs: hosted games and avatar renders.
- **Gameserver image** (`Gameserver/`) is an optional hardened container that
  each hosted game runs in: unprivileged, read-only, with memory, CPU and
  process limits. Renders stay native, because they need an X display.

## Public repositories

| Repository | What |
| --- | --- |
| [AyaEX-PUBLIC/AyaEX-MercuryCore](https://github.com/AyaEX-PUBLIC/AyaEX-MercuryCore) | The website, the Economy service, and the setup and deployment scripts |
| [AyaEX-PUBLIC/AyaEX-MercuryRCC](https://github.com/AyaEX-PUBLIC/AyaEX-MercuryRCC) | This repository |
| [AyaEX-PUBLIC/AyaEX](https://github.com/AyaEX-PUBLIC/AyaEX) | Prebuilt AyaEX clients and server, including the `Aya.Server` these services run |

These are downstream mirrors. AyaEX is developed in a private, self-hosted Git
repository, and the public repositories are updated from it, so they can be
behind it.

## Setting up

Clone this repository as `MercuryRCC`, beside `MercuryCore` (see its README),
and unpack the AyaEX server archive from the
[AyaEX releases](https://github.com/AyaEX-PUBLIC/AyaEX/releases) into
`aya-dev` beside both. You need Go, and Docker for the gameserver image.

1. Copy `Proxy/.env.example` to `Proxy/.env` and `Orbiter/.env.example` to
   `Orbiter/.env`, then fill them in:
   - `RCC_EXE` and the `--contentPath` in `RCC_ARGS` are absolute paths to
     your `aya-dev` folder.
   - `RCC_KEY` and `--httpAccessKey` are the site's `RCC_KEY` and `ASSET_KEY`.
     `GAMESERVER_KEY` is the site's too, and `PROXY_KEY` must be the same in
     both files.
   - MercuryCore's `aya/set-domain.sh <url>` puts your site's public address
     in `--baseUrl` and the `--httpAllowHost` and `--scriptHttpAllowHost`
     allowlists, so run it after creating `Proxy/.env`.
2. Optionally, build the gameserver image with `Gameserver/build.sh`. Pulling
   its Docker Hardened Images base needs `docker login dhi.io` once. To run
   games as plain processes instead, leave `GAMESERVER_IMAGE` empty.
3. Run `./run.sh` to build both services and start them in the background,
   with logs in `logs/`. `./stop.sh` stops them and every `Aya.Server` they
   started.

Run them from a desktop session, because avatar renders create an OpenGL
window and need `DISPLAY`. Restarting the services disconnects every player;
`curl localhost:64991/` lists the running gameservers.

## Credits

The original RCCService and Orbiter are by tp-link-extender and contributors
([upstream repository](https://github.com/tp-link-extender/RCCService)). This
fork keeps its [Unlicense](UNLICENSE).
