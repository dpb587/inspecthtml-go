#!/bin/bash

set -euo pipefail

if [ ! -d tmp/content-extractor-benchmark/.git ]; then
  git clone https://github.com/markusmobius/content-extractor-benchmark.git tmp/content-extractor-benchmark
fi

cd examples

find ../tmp/content-extractor-benchmark/files \
  -name '*.html' \
  -not -name 'finanzcheck.de.finanzierung.html' `# expected style mismatch` \
  -not -name 'mitvergnuegen.de.herbst.html' `# expected style mismatch` \
  -print0 \
  | sort -z \
  | xargs -0 -n1 -- go run ./dev-compare
