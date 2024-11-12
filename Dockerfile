FROM golang:1.23-bullseye as build

WORKDIR /go/src/storage

COPY go.* .
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -o /go/bin/storage ./cmd
RUN CGO_ENABLED=0 go build -o /go/bin/do-storage ./cmd/do

FROM gcr.io/distroless/static-debian12
COPY --from=build /go/bin/storage /usr/bin/
COPY --from=build /go/bin/do-storage /usr/bin/

ENTRYPOINT ["/usr/bin/storage"]