FROM golang:1.22-alpine AS builder
LABEL authors="kerrek8"

WORKDIR /usr/local/src

RUN apk --no-cache add git task

COPY ["go.mod", "go.sum", "./"]
RUN  go mod download

COPY . ./
RUN  go build -o ./bin/sso ./cmd/sso/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/sso /
COPY ./config/config.yaml /config.yaml
COPY .env /
ENV CONFIG_PATH=config.yaml
EXPOSE 44044
CMD ["/sso"]
