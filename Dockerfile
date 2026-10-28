# syntax=docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e

ARG VERSION=dev

FROM golang:1.27.3-bookworm@sha256:d9b6a1beb9474adc2a1ce842573639b44b2a0dd49b6669f51d3821baf9ec37c9 AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=linux GOARCH="${TARGETARCH:-$(go env GOARCH)}" go build -trimpath \
  -tags "osusergo,netgo" \
  -ldflags "-s -w -X github.com/prometheus/common/version.Version=${VERSION}" \
  -o /out/wakatime_exporter .

FROM gcr.io/distroless/static-debian12@sha256:d75cdd72874d4790092fcb1b058493ecf6bb5bf2b2b897045b00ff01d91843f2

COPY --from=builder --chown=nobody:nobody /out/wakatime_exporter /bin/wakatime_exporter

USER nobody
ENTRYPOINT ["/bin/wakatime_exporter"]
EXPOSE 9212
