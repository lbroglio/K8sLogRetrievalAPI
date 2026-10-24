FROM docker.io/library/golang:latest as builder

WORKDIR /app
COPY . .

RUN make test build-static

FROM alpine:latest

COPY --from=builder /app/bin/k8slogretrievalapi /usr/local/bin/k8slogretrievalapi

CMD ["k8slogretrievalapi"]