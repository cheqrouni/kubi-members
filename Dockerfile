FROM golang:latest as build
WORKDIR $GOPATH/src/github.com/ca-gip/kubi-members
COPY . $GOPATH/src/github.com/ca-gip/kubi-members
RUN make build

FROM scratch
WORKDIR /root/
COPY --from=build /go/src/github.com/ca-gip/kubi-members/build/kubi-members .
EXPOSE 8000
CMD ["./kubi-members"]