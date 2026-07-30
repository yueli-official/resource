FROM golang:1.25.12-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY api ./api
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/resource-api ./cmd/resource
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/resource-healthcheck ./cmd/healthcheck
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/resource-bindingcheck ./cmd/bindingcheck
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/resource-bootstrap ./cmd/bootstrap

FROM migrate/migrate:v4.19.1 AS migration
COPY manifest/sql/migrations /migrations

FROM gcr.io/distroless/static-debian12:nonroot AS bindingcheck
COPY --from=build /out/resource-bindingcheck /usr/local/bin/resource-bindingcheck
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/resource-bindingcheck"]

FROM gcr.io/distroless/static-debian12:nonroot AS bootstrap
COPY --from=build /out/resource-bootstrap /usr/local/bin/resource-bootstrap
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/resource-bootstrap"]

FROM gcr.io/distroless/static-debian12:nonroot AS runtime
WORKDIR /app
COPY --from=build /out/resource-api /usr/local/bin/resource-api
COPY --from=build /out/resource-healthcheck /usr/local/bin/resource-healthcheck
COPY --chown=nonroot:nonroot manifest ./manifest

ENV GF_SERVER_ADDRESS=0.0.0.0:8083
ENV OTEL_TRACES_EXPORTER=none

EXPOSE 8083
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/resource-api"]
