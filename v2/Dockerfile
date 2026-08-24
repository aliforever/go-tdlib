# go-tdlib v2 end-to-end test image.
#
# Builds the v2 login test harness (examples/login) against libtdjson.
# The host-built TDLib 1.8.64 shared library is staged in docker-libs/
# (Ubuntu 24.04 / glibc 2.39 — matches the ubuntu:24.04 base).
#
# Build (context = repository root):
#
#	docker build -f v2/Dockerfile -t go-tdlib-v2-test .
#
# Run (secrets stay in .env; entrypoint maps API_ID/API_HASH):
#
#	docker run --env-file .env -e TDLIB_PHONE_NUMBER=+1... \
#	    -v $(pwd)/.tdlib-test-data:/data go-tdlib-v2-test

FROM ubuntu:24.04 AS build

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates wget gcc g++ libssl3 zlib1g && \
    rm -rf /var/lib/apt/lists/*

RUN wget -q https://dl.google.com/go/go1.27.0.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.27.0.linux-amd64.tar.gz && \
    rm go1.27.0.linux-amd64.tar.gz
ENV PATH=/usr/local/go/bin:$PATH

COPY v2/docker-libs/lib/ /usr/local/lib/
COPY v2/docker-libs/include/ /usr/local/include/
RUN ldconfig

WORKDIR /src
COPY v2/ ./v2/
RUN cd v2 && \
    go build -tags tdjson -o /usr/local/bin/tdlib-login ./examples/login && \
    go build -tags tdjson -o /usr/local/bin/tdlib-apitest ./examples/apitest && \
    go build -tags tdjson -o /usr/local/bin/tdlib-mcp ./cmd/tdlib-mcp

FROM ubuntu:24.04

RUN apt-get update && \
    apt-get install -y --no-install-recommends libssl3 zlib1g libstdc++6 && \
    rm -rf /var/lib/apt/lists/*

COPY --from=build /usr/local/lib/libtdjson.so.1.8.64 /usr/local/lib/libtdjson.so.1.8.64
RUN ln -s libtdjson.so.1.8.64 /usr/local/lib/libtdjson.so && ldconfig
COPY --from=build /usr/local/bin/tdlib-login /usr/local/bin/tdlib-login
COPY --from=build /usr/local/bin/tdlib-apitest /usr/local/bin/tdlib-apitest
COPY --from=build /usr/local/bin/tdlib-mcp /usr/local/bin/tdlib-mcp
COPY v2/docker-libs/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["entrypoint.sh"]
