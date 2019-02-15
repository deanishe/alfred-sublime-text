
test -f alfredenv.sh && {
  source alfredenv.sh
} || {
  echo ./alfredenv.sh not found >&2
}

export GO111MODULE=on
