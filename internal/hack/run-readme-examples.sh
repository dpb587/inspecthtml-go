#!/bin/bash

set -euo pipefail

cd examples

go mod tidy

go run ./parse-dump <( echo '<p class="headline"><strong>hello</strong><br data-example />world<!-- end-->' ) \
  > parse-dump/readme-output.txt
