# Create build environment
FROM golang:1 AS build
WORKDIR /app

# Copy in source code
COPY . .
RUN cp --no-clobber config.sample.toml config.toml

# Disable "sky" TODO: Read API key from env
RUN sed -i '/\[sky\]/,/enabled = true/ s/enabled = true/enabled = false/' config.toml

# Install build deps
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
	&& apt-get install --yes \
	curl tar unzip

# Download wordnet
RUN curl -sfSL \
	'https://wordnetcode.princeton.edu/3.0/WNdb-3.0.tar.gz' \
	| tar -xzf - -C data && mv data/dict data/wordnet

# Download cities
RUN curl -sfSLO \
	'https://download.geonames.org/export/dump/cities15000.zip' \
	&& unzip -o 'cities15000.zip' -d data && rm 'cities15000.zip'

# Download IFSC codes (v2.0.54)
RUN mkdir data/ifsc && curl -sfSL \
	'https://github.com/razorpay/ifsc/releases/download/v2.0.54/by-bank.tar.gz' \
	| tar -xzf - --strip-components=1 -C data/ifsc

# Make binary
RUN make build

# Runtime
FROM alpine:3
WORKDIR /app
COPY --from=build /app /app
CMD ["./dnstoys.bin"]
