FROM golang:1.27-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/fortiweb_exporter ./cmd/fortiweb_exporter

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/fortiweb_exporter /fortiweb_exporter

EXPOSE 9633

ENTRYPOINT ["/fortiweb_exporter"]
CMD ["--config", "/etc/fortiweb_exporter/config.yml"]
