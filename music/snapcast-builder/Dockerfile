# Install SnapServer on minimal OS - script v4.0.1 [2022-03-30]
ARG ALPINE_BASE="3.15"

# SnapCast build stage
FROM alpine:${ALPINE_BASE} as builder
RUN apk -U add \
    alsa-lib-dev \
    avahi-dev \
    bash \
    boost-dev \
    build-base \
    ccache \
    cmake \
    expat-dev \
    flac-dev \
    git \
    libvorbis-dev \
    opus-dev \
    soxr-dev

ARG VERSION=master
RUN <<EOF 
    git clone --recursive --depth 1 --branch $VERSION https://github.com/badaix/snapcast.git
    cd snapcast

    cmake -S . -B build \
        -DCMAKE_CXX_COMPILER_LAUNCHER=ccache \
        -DBUILD_WITH_PULSE=OFF \
        -DCMAKE_BUILD_TYPE=Release \
        ..
    cmake --build build --parallel 3
EOF