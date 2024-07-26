FROM golang:1.22.2-alpine3.19 AS builder

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN apk add alpine-sdk

RUN GO111MODULE=on CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -tags musl -o app -v ./cmd/app/main.go

FROM golang:1.22.2-alpine3.19

ARG NO_ROOT_USER="app"

RUN apk --no-cache add ca-certificates && \
    addgroup -S ${NO_ROOT_USER} && adduser -S ${NO_ROOT_USER} -G ${NO_ROOT_USER} -h /home/${NO_ROOT_USER}

WORKDIR /home/${NO_ROOT_USER}

COPY --from=builder /src/app .

USER ${NO_ROOT_USER}

EXPOSE 8080

ENTRYPOINT ["/home/app/app"]
