FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /spooltotag .

FROM alpine:3.20
COPY --from=build /spooltotag /spooltotag
EXPOSE 8080
ENTRYPOINT ["/spooltotag"]
