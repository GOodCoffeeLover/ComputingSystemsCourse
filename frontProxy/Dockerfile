FROM golang:1.18 AS base

WORKDIR /usr/src/app

COPY go.mod go.sum ./

RUN ["go", "mod", "download"]

COPY main.go .

RUN ["go", "build", "-o", "/proxy", "main.go"]

FROM golang:1.18

WORKDIR frontProxy

EXPOSE 8080:8080

COPY --from=base /proxy .

CMD ["./proxy"]
