##
## Build
##

#FROM golang:1.23-bookworm AS build
FROM public.ecr.aws/docker/library/golang:1.23-bookworm AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o service main.go

##
## Deploy
##

FROM gcr.io/distroless/static-debian12:latest

WORKDIR /

COPY --from=build /app/service /app/*config.json ./

EXPOSE 3000

USER nonroot:nonroot

ENTRYPOINT ["/service"]