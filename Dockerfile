# docker build -t go-account .
# docker run -dit --name go-account -p 5000:5000 go-account
# docker run -dit --name go-account-v1 -p 80:5000 908671954593.dkr.ecr.us-east-2.amazonaws.com/go-account-v1:latest

FROM golang:1.24 As builder

RUN apt-get update && apt-get install bash && apt-get install -y --no-install-recommends ca-certificates

WORKDIR /app
COPY . .
RUN go mod tidy

WORKDIR /app/cmd
RUN go build -o go-account -ldflags '-linkmode external -w -extldflags "-static"'

FROM alpine

WORKDIR /app
COPY --from=builder /app/cmd/go-account .

WORKDIR /var/pod/secret
RUN echo -n "postgres" > /var/pod/secret/username
RUN echo -n "postgres" > /var/pod/secret/password

COPY --from=builder /app/cmd/.env .

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app/go-account"]