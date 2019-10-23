# ------------------------------------------------------------------------
from mcluseau/golang-build:1.13.3

# ------------------------------------------------------------------------
from alpine:3.10
run apk add --update mksquashfs

add layer/ /layer/
copy --from=build /go/bin/ /layer/sbin/
