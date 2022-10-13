FROM golang:1.18

WORKDIR main

EXPOSE 8080:8080

# COPY ./wait-for-it.sh .
# RUN ["chmod", "+x", "./wait-for-it.sh"]

COPY go.mod go.sum ./

RUN ["go", "mod", "download"]

COPY ./cmd ./cmd
COPY ./internal ./internal

RUN ["go", "build", "-o", "/task-manager", "cmd/main.go"]

# CMD ["./wait-for-it.sh", "-t", "0", "mongoDB:27017", "--",  "/task-manager"]
CMD ["/task-manager"]
