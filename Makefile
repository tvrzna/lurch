DIST_FILE=lurch

clean:
	rm -r dist

build:
	mkdir -p dist
	go build -o dist/${DISTFILE} -buildvcs=false

install:
	install -DZs dist/${DISTFILE} ${DESTDIR}/usr/bin
