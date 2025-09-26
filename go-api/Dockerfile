# build
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/auditor ./cmd/auditor

# run
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /out/auditor /app/auditor
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/api"]