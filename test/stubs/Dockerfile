ARG BASE_CARDS_IMAGE
ARG BASE_IMAGE
# -----------------------------------------------
# Stage1: Stubs Builder
# -----------------------------------------------
FROM ${BASE_CARDS_IMAGE} AS builder

RUN --mount=type=cache,target=/root/.cache/go-build,mode=0777 \
    GOOS=linux go build -o ./dist/stubs ./test/stubs

# -----------------------------------------------
# Stage2: Stubs Runtime
# -----------------------------------------------
FROM ${BASE_IMAGE}

COPY --from=builder /src/dist/stubs /bin
COPY --from=builder /src/test/data /test/data

CMD ["/bin/stubs"]
