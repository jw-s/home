# Install SnapServer on minimal OS - script v4.0.1 [2022-03-30]
ARG ALPINE_BASE="3.15"
# Final stage
FROM alpine:${ALPINE_BASE}

RUN apk add --no-cache \
        alsa-lib \
        avahi-libs \
        expat \
        flac \
        libvorbis \
        opus \
        soxr

COPY --from=joelws/snapcast-builder snapcast/bin/snapclient /usr/local/bin/

EXPOSE 1704
EXPOSE 1705
EXPOSE 1780

ENTRYPOINT ["snapclient"]