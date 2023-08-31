FROM golang:1.21-alpine

WORKDIR /

ARG create_tables=false

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -o main ./cmd/web

EXPOSE 8080

CMD [ "./main", "-create_tables", "${create_tables}" ]