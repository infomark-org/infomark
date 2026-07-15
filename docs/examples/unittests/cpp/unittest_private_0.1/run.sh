#!/usr/bin/env sh

rm -rf build || true
mkdir build
cd build
cmake ..
make
./hello -s