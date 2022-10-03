FROM golang:1.18


WORKDIR ComputingSystemsCourse
#ENV GOROOT=/ComputingSystemsCourse
EXPOSE 8080:8080

COPY go.mod .
COPY go.sum .

RUN ["go", "mod", "download"]

COPY ./wait-for-it.sh .
RUN ["chmod", "+x", "./wait-for-it.sh"]
COPY cmd ./cmd
COPY internal ./internal

RUN ["go", "build", "-o", "/task-manager", "cmd/main.go"]

CMD ["./wait-for-it.sh", "-t", "0", "mongoDB:27017", "--",  "/task-manager"]