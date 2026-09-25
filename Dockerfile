FROM golang:1.25.0 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN GOPROXY=https://goproxy.cn,direct go mod download

COPY . .

RUN go build -o naotodoserver cmd/main.go

FROM golang:1.25.0

WORKDIR /app

COPY --from=builder /app/naotodoserver .
COPY --from=builder /app/conf ./conf
COPY --from=builder /app/infrastructure/ip2region ./infrastructure/ip2region

EXPOSE 443

CMD ["./naotodoserver"]