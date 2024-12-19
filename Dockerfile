FROM alpine:latest

# Update the package list and install Go
RUN apk update && apk add --no-cache go git

COPY ./ ./
RUN go install ./cmd/manifest

CMD git diff | go run cmd/manifest/main.go inspect