FROM node:16-alpine as build-front

COPY . /usr/src/qa/front/
WORKDIR /usr/src/qa/front/examples/qa/assets/flagship-qa-front

RUN npm install
RUN npm run build-bundle

FROM golang:1-alpine as build-env

COPY . /go/src/github.com/flagship-io/flagship-go-sdk
WORKDIR /go/src/github.com/flagship-io/flagship-go-sdk/examples/qa

# Get dependancies - will also be cached if we won't change mod/sum
RUN go mod download

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/example *.go

# Run the binary
FROM alpine

EXPOSE 8080

WORKDIR /go/bin

COPY --from=build-env /go/bin/example example
COPY --from=build-front /usr/src/qa/front/examples/qa/assets/ qa/assets/

CMD ["./example"]