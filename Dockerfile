FROM docker.io/library/golang:1.16-bullseye as builder
RUN go get -d -v github.com/google/go-github/github && \
    go get -d -v github.com/joho/godotenv && \
    go get -d -v golang.org/x/oauth2

FROM docker.io/library/golang:1.16-bullseye
COPY --from=builder /go /go
WORKDIR /go/src/fiskil
COPY publisher publisher
RUN go test -v ./...
# RUN go install -v
# CMD ["publisher"]
