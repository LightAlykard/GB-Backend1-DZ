FROM golang:1.16 as modules

ADD go.mod go.sum /m/
RUN cd /m && go mod download

FROM golang:1.16 as builder

COPY --from=modules /go/pkg /go/pkg

RUN useradd -u 12345 shortener_user

RUN mkdir -p /app
ADD . /app
WORKDIR /app

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o ./bin/shortener ./cmd/main.go

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
USER shortener_user

COPY --from=builder /app/bin/shortener /shortener

CMD ["/shortener"]
