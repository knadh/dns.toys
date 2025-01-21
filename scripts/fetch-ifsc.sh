#!/bin/bash

# change data path as required
cd ../data
mkdir -p ifsc
cd ifsc
mkdir -p by-bank
rm -r by-bank
curl -L -s $(curl -s https://api.github.com/repos/razorpay/ifsc/releases/latest | grep "browser_download_url" | grep "by-bank.tar.gz" | sed -E 's/.*"([^"]+)".*/\1/') -o by-bank.tar.gz
tar -xzf by-bank.tar.gz
rm by-bank.tar.gz
mv by-bank/*.json .
rm -r by-bank