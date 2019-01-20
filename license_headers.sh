#!/bin/bash

set -e

update-license-header() {
    perl -i -0pe 's/\/\/ Copyright .* Stratumn.*\n(\/\/.*\n)*/`cat LICENSE_HEADER`/ge' $1
}

for f in $(find . -name "*.go" | grep -v vendor | grep -v ui/); do
    echo Updating $f...
    update-license-header $f
done

for f in $(find ui -name "*.tsx" | grep -v node_modules | grep -v build); do
    echo Updating $f...
    update-license-header $f
done

for f in $(find ui -name "*.ts" | grep -v node_modules | grep -v build); do
    echo Updating $f...
    update-license-header $f
done
