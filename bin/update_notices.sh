#!/bin/bash

set -e

go install github.com/google/go-licenses@latest
go-licenses check ./... --allowed_licenses="MIT,Apache-2.0,BSD-3-Clause,MPL-2.0,BSD-2-Clause"
go-licenses report ./... --template notices.tpl > NOTICES.txt

set +e