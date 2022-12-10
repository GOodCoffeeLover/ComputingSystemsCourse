FROM golang:1.18

WORKDIR calculation

EXPOSE 8090:8090


COPY go.mod go.sum ./

RUN ["go", "mod", "download"]

COPY ./cmd ./cmd
COPY ./internal ./internal
COPY ./pkg ./pkg


RUN ["go", "build", "-o", "calculator", "cmd/main.go"]

CMD ./calculator