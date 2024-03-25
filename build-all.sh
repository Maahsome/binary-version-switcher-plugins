#!/usr/bin/env bash

SEMVER=v0.0.1
for d in $(ls -1 -d */); do
  pushd ${d}
  ./build.sh
  popd
done

cp **/*.so ~/.config/binary-version-switcher/plugins/${SEMVER}/

