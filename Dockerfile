# TODO: Move to cmd/app

FROM golang:1.16

WORKDIR /app

COPY /go.mod ./
COPY /go.sum ./
RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal

RUN go build -o /bank-service ./cmd/app

# TODO: Make multistage

EXPOSE 8080

CMD [ "/bank-service" ]