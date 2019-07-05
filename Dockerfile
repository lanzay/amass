FROM golang:alpine as build
WORKDIR /go/src/github.com/OWASP/Amass
COPY . .
RUN apk --no-cache add git \
  && go get -u github.com/lanzay/Amass/...

FROM alpine:latest
COPY --from=build /go/bin/amass /bin/amass
COPY --from=build /go/src/github.com/lanzay/Amass/wordlists /wordlists
ENTRYPOINT ["/bin/amass"]
