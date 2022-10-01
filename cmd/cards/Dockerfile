ARG BASE_CARDS_IMAGE
ARG BASE_RUNTIME_IMAGE
# -----------------------------------------------
# Stage1: Cards Builder
# -----------------------------------------------
FROM ${BASE_CARDS_IMAGE} AS builder

ARG MONITORING_PKG=github.com/anzx/pkg/monitoring
ARG HEALTH_PKG=github.com/anz-bank/pkg/health
ARG CARDS_PKG=github.com/anzx/fabric-cards/pkg

ARG VERSION
ARG REPOURL
ARG COMMITHASH
ARG BUILDLOGURL
ARG CONTAINERTAG

# Link flags for cards
ARG LDFLAGS="${LDFLAGS} -X ${CARDS_PKG}/util/initutil.Version=${VERSION}"
ARG LDFLAGS="${LDFLAGS} -X ${CARDS_PKG}/util/initutil.ApplicationName=Cards"

# Link flags for monitoring
ARG LDFLAGS="${LDFLAGS}${COMMITHASH:+ -X ${MONITORING_PKG}/attributes.GitHash=${COMMITHASH}}"
ARG LDFLAGS="${LDFLAGS}${REPOURL:+ -X ${MONITORING_PKG}/attributes.RepoURL=${REPOURL}}"
# value should match the WORKDIR
ARG LDFLAGS="${LDFLAGS} -X ${MONITORING_PKG}/attributes.BuildPath=/src"

# Link flags for health
ARG LDFLAGS="${LDFLAGS}${REPOURL:+ -X ${HEALTH_PKG}.RepoURL=${REPOURL}}"
ARG LDFLAGS="${LDFLAGS}${COMMITHASH:+ -X ${HEALTH_PKG}.CommitHash=${COMMITHASH}}"
ARG LDFLAGS="${LDFLAGS}${BUILDLOGURL:+ -X ${HEALTH_PKG}.BuildLogURL=${BUILDLOGURL}}"
ARG LDFLAGS="${LDFLAGS}${CONTAINERTAG:+ -X ${HEALTH_PKG}.ContainerTag=${CONTAINERTAG}}"
ARG LDFLAGS="${LDFLAGS}${SEMVER:+ -X ${HEALTH_PKG}.Semver=${VERSION}}"

ARG LDFLAGS="${LDFLAGS} -linkmode=external"

RUN --mount=type=cache,target=/root/.cache/go-build,mode=0777 GOOS=linux  \
    go build -ldflags="${LDFLAGS}" -o ./dist/cards ./cmd/cards

# -----------------------------------------------
# Stage2: Cards Runtime
# -----------------------------------------------
FROM ${BASE_RUNTIME_IMAGE}

COPY --from=builder /src/dist/cards /bin

LABEL ci_group="ANZx-Fabric"
LABEL ci_name="ANZx-Platform"

EXPOSE 8080

CMD ["/bin/cards"]
