FROM golang:1.19-alpine as build
# Set the working directory
WORKDIR /go/src/axios-be-exercise-kenston-oneal
# Copy and download dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download
# Copy the source files from the host
COPY . /go/src/storymetadatagenerator

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/storymetadatagenerator ./cmd/storymetadatagenerator

FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=build /go/bin/storymetadatagenerator /bin/storymetadatagenerator
ENTRYPOINT ["/bin/storymetadatagenerator", "--config-file=/etc/config.json"]
