#!/usr/bin/env bash

wget https://github.com/transhift/transhift/releases/download/v0.1.0-alpha/transhift-v0.1.0-alpha.tar.gz
tar -xf transhift-v0.1.0-alpha.tar.gz
rm transhift-v0.1.0-alpha.tar.gz
sudo mv transhift /usr/bin
