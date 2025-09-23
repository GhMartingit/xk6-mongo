#!/usr/bin/env bats

setup() {
  BASEDIR="$(git rev-parse --show-toplevel)"
  cd $BASEDIR

  EXE="go run ./tools/docsme"
}

@test "run docsme (stdout)" {
  run $EXE
  [ $status -eq 0 ]
  echo "$output" | grep -q 'Keep documentation up to date based on CLI help'
}

@test "run docsme (file)" {
  FILE=$BATS_TEST_TMPDIR/README.md
  cp README.md $FILE
  $EXE -r cli -o $FILE
  grep -q 'Keep documentation up to date based on CLI help' $FILE
}
