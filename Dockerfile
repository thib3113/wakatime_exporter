# syntax=docker/dockerfile:1

ARG VERSION=dev

FROM golang:1.27.2-bookworm@sha256:5cf287a799e6b94384bad13d16b14904c531f51ba65792237e122ce42b392f61 AS builder

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
