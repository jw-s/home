# Install SnapServer on minimal OS - script v4.0.1 [2022-03-30]
ARG ALPINE_BASE="3.15"

# SnapWeb build stage
FROM node:alpine as snapweb

RUN <<EOF
    apk -U add build-base git
    npm install -g typescript
    git clone https://github.com/badaix/snapweb
    make -C snapweb
EOF

# Final stage
FROM alpine:${ALPINE_BASE}
LABEL maintainer="Saiyato"

RUN <<EOF
    apk add --no-cache \
        alsa-lib \
        avahi-libs \
        expat \
        flac \
        libvorbis \
        opus \
        soxr \
        alsa-utils
EOF

COPY --from=joelws/snapcast-builder snapcast/bin/snapserver /usr/local/bin/
COPY --from=snapweb snapweb/dist/ /usr/share/snapserver/snapweb

EXPOSE 1704
EXPOSE 1705
EXPOSE 1780

ENTRYPOINT ["snapserver"]