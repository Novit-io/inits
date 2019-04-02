# ------------------------------------------------------------------------
from golang:1.12.0 as build

env CGO_ENABLED=0

workdir /src
add go.mod go.sum /src/
run go mod download

add . ./
run go test ./...
run go install ./cmd/...

# ------------------------------------------------------------------------
from alpine:3.9
run apk add --update mksquashfs

add layer/ /layer/
copy --from=build /go/bin/ /layers/sbin/
