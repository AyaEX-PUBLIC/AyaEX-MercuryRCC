# syntax=docker/dockerfile:1
# Hardened runtime for hosted Aya.Server gameservers, on Docker Hardened
# Images (Debian 13, https://www.docker.com/products/hardened-images/).
#
# The dev image installs the libraries Aya.Server links against, and only the
# ones it actually loads are copied into the minimal runtime image, which runs
# as an unprivileged user. Aya.Server, its content and the DataStore are
# mounted by the RCC proxy when it starts a gameserver (Proxy/RCCService.go),
# so a new Aya build needs no new image, unless it links a new library.
# Build with ./build.sh, after `docker login dhi.io` once.

ARG BASE=dhi.io/debian-base:trixie
ARG DEV=dhi.io/debian-base:trixie-dev

FROM ${DEV} AS libs
SHELL ["/bin/bash", "-euo", "pipefail", "-c"]
# the Aya.Server build directory, a named build context (build.sh)
COPY --from=aya Aya.Server libfmod.so.13 /aya/
# (init-system-helpers first: x11-common's install script needs update-rc.d)
RUN apt-get update \
	&& apt-get install -y --no-install-recommends init-system-helpers \
	&& apt-get install -y --no-install-recommends \
		libatomic1 libglu1-mesa libglx0 libice6 libopengl0 libpulse0 \
		libsm6 libstdc++6 libuuid1 libx11-6 libxext6 \
	&& rm -rf /var/lib/apt/lists/*
# Copy each library Aya.Server loads, except the C library the runtime image
# already has, keeping its directory (as /usr/lib, Debian's merged /usr)
RUN mkdir -p /out/usr/lib/x86_64-linux-gnu \
	&& cp /aya/libfmod.so.13 /out/usr/lib/x86_64-linux-gnu/ \
	&& if LD_LIBRARY_PATH=/aya ldd /aya/Aya.Server | grep "not found"; then exit 1; fi \
	&& LD_LIBRARY_PATH=/aya ldd /aya/Aya.Server \
		| awk '$2 == "=>" && $3 ~ /^\// { print $3 }' \
		| while read -r lib; do \
			case "$(basename "$lib")" in \
				libc.so.* | libm.so.* | ld-linux* | libpthread.so.* | libdl.so.* | librt.so.* | libfmod.so.*) continue ;; \
			esac; \
			dir="$(dirname "$lib" | sed 's#^/lib/#/usr/lib/#')"; \
			mkdir -p "/out$dir"; \
			cp -L "$lib" "/out$dir/"; \
		done

FROM ${BASE}
COPY --from=libs /out/ /
# Aya.Server reads shaders/, platform/ and ClientSettings.json from its directory
WORKDIR /aya
ENTRYPOINT ["/aya/Aya.Server"]
