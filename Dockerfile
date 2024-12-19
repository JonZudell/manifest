FROM alpine:latest
RUN adduser -D -u 1001 manifestuser
WORKDIR /manifest
# Update the package list and install Go
RUN apk update && apk add --no-cache go git

COPY ./ ./
USER manifestuser

RUN go install -buildvcs=false ./cmd/manifest
ENV PATH="/home/manifestuser/go/bin:${PATH}"
WORKDIR /app

CMD git diff | manifest inspect --formatter github