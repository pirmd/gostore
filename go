#!/bin/sh

# go.sh is a simple wrapper around go to override some operation, for example to enforce some ldflags

PROG=${PWD##*/}
VERSION=$(git describe --tags 2>/dev/null)
BUILD=$(git rev-parse --short HEAD 2>/dev/null)
MANDIR=$GOBIN/../man/ #on linux probably prefer $GOBIN/../share/man/

if [ $# -gt 0 ]; then CMD=$1; shift; fi

case ${CMD:-} in
    (build)
        dirs=$(go list -f {{.Dir}} ./...)
        for d in $dirs; do goimports -w $d/*.go; done
        #Nota: if another ldflags directive is given from the command line, it
        #will override this directive, it can be nice to have if you like
        #another set-up for initiliazing version variables but can be annoying
        #in other cases. For these cases, you'd better edit this script
        #directly
        exec go build -ldflags "-X github.com/pirmd/clapp.version=${VERSION:-?.?.?} -X github.com/pirmd/clapp.build=${BUILD:-?}" "$@"
    ;;

    (install)
        #generate and install manpages
        go generate
        install -m 644 *.1 $MANDIR/man1/

        #Nota: if another ldflags directive is given from the command line, it
        #will override this directive, it can be nice to have if you like
        #another set-up for initiliazing version variables but can be annoying
        #in other cases. For these cases, you'd better edit this script
        #directly
        exec go install -ldflags "-s -w -X github.com/pirmd/clapp.version=${VERSION:-?.?.?} -X github.com/pirmd/clapp.build=${BUILD:-?}" "$@"
    ;;

    (release)
        tag=${1?Provide a tag name (ex: $0 release v1.0.0)}
        git tag $tag
        git archive --prefix "${PROG}-$tag/" "$tag" -o "./snapshots/$(date +%Y%m%d)_${PROG}-$tag.tar.gz"
    ;;

    (*)
        exec go $CMD "$@"
	;;
esac

