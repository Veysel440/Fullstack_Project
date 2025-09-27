# ---- build
FROM golang:1.23-alpine AS build
WORKDIR /src
RUN apk add --no-cache build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN go build -o /out/api            ./cmd/api && \
    go build -o /out/auditor        ./cmd/auditor && \
    go build -o /out/auditor-retry  ./cmd/auditor-retry

# ---- runtime
FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/api /api
COPY --from=build /out/auditor /auditor
COPY --from=build /out/auditor-retry /auditor-retry
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
