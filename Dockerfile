FROM golang:latest

WORKDIR usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build .

CMD ["./liveness-sim-test", "-n", "4", "-t"]
