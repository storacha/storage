FROM golang:1.23-bullseye as build

WORKDIR /go/src/storage

COPY go.* .
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -o /go/bin/storage ./cmd

FROM gcr.io/distroless/static-debian12
COPY --from=build /go/bin/storage /usr/bin/

ENTRYPOINT ["/usr/bin/storage"]