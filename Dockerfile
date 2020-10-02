# ------------------------------------------------------------------------
from mcluseau/golang-builder:1.15.1 as build

# ------------------------------------------------------------------------
from alpine:3.12
run apk add --update squashfs-tools

add layer/ /layer/
copy --from=build /go/bin/ /layer/sbin/
