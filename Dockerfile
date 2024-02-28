FROM golang:1.22 AS build-go
ENV CGO_ENABLED=0
ARG BUILD_VERSION
COPY . /app
WORKDIR /app
RUN go build -o drone -ldflags "-X github.com/m-mizutani/drone/pkg/domain/types.AppVersion=${BUILD_VERSION}" .

FROM gcr.io/distroless/base
COPY --from=build-go /app/drone /drone

ENTRYPOINT ["/drone"]
