FROM golang:1.17 as builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o mc-monitor

FROM scratch
ENTRYPOINT ["/mc-monitor"]
COPY --from=builder /build/mc-monitor /mc-monitor
