FROM alpine:latest

WORKDIR /manifest
# Update the package list and install Go
RUN apk update && apk add --no-cache go git

COPY ./ ./

RUN go install ./cmd/manifest && go build -o manifest cmd/manifest/main.go

WORKDIR /app

CMD git diff | go run /manifest/cmd/manifest/main.go inspect --formatter github --strict