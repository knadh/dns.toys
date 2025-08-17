#!/bin/bash

if [ -z "$1" ]; then
    echo "Usage: $(basename $0) ./path/ifsc"
    exit 1
fi

URL=$(curl -s https://api.github.com/repos/razorpay/ifsc/releases/latest | grep "browser_download_url" | grep "by-bank.tar.gz" | sed -E 's/.*"([^"]+)".*/\1/')

echo "downloading $URL ==> $1"
mkdir -p "$1"
curl -L "$URL" | tar -xz --strip-components=1 -C "$1"
