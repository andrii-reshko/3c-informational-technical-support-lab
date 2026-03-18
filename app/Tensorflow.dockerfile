FROM alpine:3.21 AS loader

RUN apk add --no-cache wget tar

WORKDIR /tflite

RUN wget https://github.com/mattn/go-tflite/releases/download/v1.0.5/go-tflite-buildkit-20240529.tar.gz && \
    mkdir -p extracted && \
    tar -C extracted -xvf go-tflite-buildkit-20240529.tar.gz

FROM scratch AS artifacts
COPY --from=loader /tflite/extracted/lib/libtensorflowlite_c.so /lib/
COPY --from=loader /tflite/extracted/include/tensorflow /include/tensorflow