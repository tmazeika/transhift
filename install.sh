#!/usr/bin/env bash

wget https://github.com/transhift/transhift/releases/download/v0.1.0-alpha/transhift-0.1.0.tar.gz
tar -xf transhift-0.1.0.tar.gz
rm transhift-0.1.0.tar.gz
sudo mv transhift /usr/bin
