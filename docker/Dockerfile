FROM alpine:latest AS helper

RUN ["apk", "add", "--no-cache", "ca-certificates"]

FROM scratch

COPY --from=helper /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY transferwise-exchange-rate /

ENTRYPOINT ["/transferwise-exchange-rate"]
