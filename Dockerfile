FROM golang:1.27-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/fortiweb-exporter ./cmd/fortiweb-exporter

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/fortiweb-exporter /fortiweb-exporter

EXPOSE 9633

ENTRYPOINT ["/fortiweb-exporter"]
CMD ["--config", "/etc/fortiweb-exporter/config.yml"]
