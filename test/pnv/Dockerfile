ARG BASE_CARDS_IMAGE
ARG BASE_IMAGE
# -----------------------------------------------
# Stage1: PNV Builder
# -----------------------------------------------
FROM ${BASE_CARDS_IMAGE} AS builder

RUN --mount=type=cache,target=/root/.cache/go-build,mode=0777 GOOS=linux \
    go build --tags=integration,pnv -o ./dist/pnv-cards ./test/pnv/cards/main.go


# -----------------------------------------------
# Stage2: PNV Runtime
# -----------------------------------------------
FROM ${BASE_IMAGE}

COPY --from=builder /src/test/common /common
COPY --from=builder /src/test/pnv/config /config
COPY --from=builder /src/dist/pnv-cards /bin/pnv-cards

WORKDIR /src
