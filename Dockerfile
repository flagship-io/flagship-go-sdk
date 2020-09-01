FROM golang:1-alpine as build-env

WORKDIR /go/src/github.com/abtasty/flagship-go-sdk

# COPY the source code as the last step
COPY . .

WORKDIR /go/src/github.com/abtasty/flagship-go-sdk/examples

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/example qa/*.go

# Run the binary
FROM alpine

EXPOSE 8080

WORKDIR /go/bin

COPY --from=build-env /go/bin/example example
COPY --from=build-env /go/src/github.com/abtasty/flagship-go-sdk/examples/qa/assets qa/assets

CMD ["./example"]