FROM golang:1.23

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

ENTRYPOINT ["/bin/bash", "-c"]