FROM golang:1.17-alpine as builder
WORKDIR $GOPATH/src/go.k6.io/k6
ADD . .
RUN CGO_ENABLED=0 go install -a -trimpath -ldflags "-s -w -X go.k6.io/k6/lib/consts.VersionDetails=$(date -u +"%FT%T%z")/$(git describe --always --long --dirty)"
RUN go install -trimpath go.k6.io/xk6/cmd/xk6@latest
RUN --mount=type=ssh xk6 build \ 
--with github.com/mostafa/xk6-kafka@latest \
--with github.com/GhMartingit/xk6-mongo
RUN cp k6 $GOPATH/bin/k6

FROM alpine:3.14
RUN apk add --no-cache ca-certificates && \
    adduser -D -u 12345 -g 12345 k6
COPY --from=builder /go/bin/k6 /usr/bin/k6

USER 12345

ENTRYPOINT ["k6-2-ext"]