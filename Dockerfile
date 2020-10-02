# ------------------------------------------------------------------------
from mcluseau/golang-builder:1.15.1 as build

# ------------------------------------------------------------------------
from alpine:3.12

entrypoint mksquashfs /layer /layer.squashfs >&2 && base64 </layer.squashfs

run apk add --update squashfs-tools

add layer/ /layer/
copy --from=build /go/bin/ /layer/sbin/
