FROM golang:1.15-alpine as build
WORKDIR /go/src/github.com/fugiman/api.clickquest.net
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o main main.go

FROM scratch
COPY --from=build  /etc/ssl/certs/ca-certificates.crt                  /etc/ssl/certs/
COPY --from=build  /go/src/github.com/fugiman/api.clickquest.net/main  /server
CMD ["/server"]
