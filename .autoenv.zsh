
test -f alfredenv.sh && {
  source alfredenv.sh
} || {
  echo ./alfredenv.sh not found >&2
}
