FROM golang:1.17.7 as builder
ARG GOPROXY
ARG APP
ENV GOPROXY=${GOPROXY}
WORKDIR /go/soc-eventer
COPY . .
RUN CGO_ENABLED=0 go build -a -o ./dist/${APP} main.go

FROM core.harbor1.domain/soc/soc-tool:v5.0.5
ARG APP
WORKDIR /
COPY --from=builder /go/soc-eventer/dist/${APP} /usr/local/bin/${APP}
USER root:root