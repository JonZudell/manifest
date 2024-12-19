FROM alpine:latest
RUN adduser -D -u 1001 manifestuser
USER manifestuser
WORKDIR /manifest
# Update the package list and install Go
RUN apk update && apk add --no-cache go git

COPY ./ ./

RUN go install ./cmd/manifest
ENV PATH="/root/go/bin:${PATH}"
WORKDIR /app

CMD git diff | go run /manifest/cmd/manifest/main.go inspect --formatter github