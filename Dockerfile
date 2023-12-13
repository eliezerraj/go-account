#docker build -t go-account .
#docker run -dit --name go-account -p 5000:5000 go-account

FROM golang:1.21 As builder

WORKDIR /app
COPY . .

WORKDIR /app/cmd
RUN go build -o go-account -ldflags '-linkmode external -w -extldflags "-static"'

FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/go-account .

CMD ["/app/go-account"]