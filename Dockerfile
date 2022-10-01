ARG BASE_IMAGE

FROM ${BASE_IMAGE} as builder

ENV GOPROXY="https://platform-gomodproxy.services-platdev.x.gcpnp.anz,https://artifactory.gcp.anz/artifactory/api/go/go,direct"
ENV GONOSUMDB="github.com/anzx/*,github.service.anz/*"
ENV GO111MODULE=on
ENV CGO_ENABLED=1

WORKDIR /src

ADD go.mod .
ADD go.sum .

RUN go mod download

ADD . .
