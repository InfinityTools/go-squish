#!/bin/sh

# A build script that automatically picks the right library from the subfolders in "libs".
# Use this script if you are unable or don't want to use the system library.

show_help() {
  echo "Usage $0 [options]"
  echo ""
  echo "Options:"
  echo "  --libdir path   Override library path"
  echo "  --help          This help"
  exit 0
}

# show_message(msg: string, level: int): Print "msg" and exit with "level". Default level is 0.
show_message() {
  if test $# != 0; then
    echo $1
    shift
  fi
  if test $# != 0; then
    exit $1
  else
    exit 0
  fi
}


# Evaluating command line arguments...
while test $# != 0
do
  case $1 in
  --libdir)
    shift
    if test $# == 0; then
      echo "Missing argument: --libdir"
      exit 1
    fi
    libdir="$1"
    ;;
  --help)
    show_help
    ;;
  esac
  shift
done


if test -z "$libdir"; then
  if test $(go env GOOS) = "darwin"; then
    libos="darwin"
    # Package-specific libraries
    ldargs="-lsquish -lm -lstdc++"
  else
    libos="linux"
    # Package-specific libraries
    ldargs="-lsquish -lgomp -lm -lstdc++"
  fi

  if test $(go env GOARCH) = "amd64"; then
    libarch="amd64"
  else
    libarch="386"
  fi
  echo "Detected: os=$libos, arch=$libarch"
  libdir=libs/$libos/$libarch
else
    echo "Using libdir: $libdir"
fi

if test ! -d $libdir; then
  echo "Error: Path does not exist: $libdir"
  exit 1
fi

echo "Building library..."
CGO_LDFLAGS="-L$libdir $ldargs" go build && go install && show_message "Finished." 0 || show_message "Cancelled." 1
