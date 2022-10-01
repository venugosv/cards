ARG BASE_CARDS_IMAGE
ARG BASE_IMAGE
# Stage1: Builder Image
FROM $BASE_CARDS_IMAGE as builder

RUN --mount=type=cache,target=/root/.cache/go-build,mode=0777 CGO_ENABLED=0 GOOS=linux \
    go test -tags=integration ./test/integration/callback -c -o integration-callback.test && \
    go test -tags=integration ./test/integration/cardcontrols -c -o integration-cardcontrols.test && \
    go test -tags=integration ./test/integration/cards -c -o integration-cards.test

# Stage2: Runtime Image
FROM ${BASE_IMAGE}

COPY --from=builder /src/test/common /common
COPY --from=builder /src/test/integration/config /config
COPY --from=builder /src/integration-callback.test /bin
COPY --from=builder /src/integration-cardcontrols.test /bin
COPY --from=builder /src/integration-cards.test /bin

WORKDIR /
