FROM golang:1.18

WORKDIR main

EXPOSE 8085:8085

COPY go.mod go.sum ./

RUN ["go", "mod", "download"]

COPY ./cmd ./cmd
COPY ./internal ./internal

RUN ["go", "build", "-o", "/task-manager", "cmd/main.go"]

# CMD ["./wait-for-it.sh", "-t", "0", "mongoDB:27017", "--",  "/task-manager"]
CMD ["/task-manager"]
