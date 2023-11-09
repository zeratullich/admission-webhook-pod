# Build the webhook-pod binary
FROM golang:1.20.10 as builder

WORKDIR /workspace

# copy source file
COPY . ./

# set golang env & build binary
RUN go env -w GOPROXY=https://goproxy.cn,direct GOSUMDB=off GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
 && go build -a -o admission-webhook-pod .

FROM alpine:latest

WORKDIR /

# install binary
COPY --from=builder /workspace/admission-webhook-pod .

ENTRYPOINT [ "/admission-webhook-pod" ]
