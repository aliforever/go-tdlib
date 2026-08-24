# docker-libs — staged TDLib binaries (not committed)

The v2 test Dockerfile (`v2/Dockerfile`) does not build TDLib from source;
it copies a prebuilt `libtdjson` and its C headers from this directory.
Everything here except this README is gitignored.

## Required layout

```
docker-libs/
├── lib/
│   ├── libtdjson.so.1.8.64   # the real shared library
│   └── libtdjson.so          # symlink -> libtdjson.so.1.8.64
├── include/
│   └── td/
│       └── telegram/         # td_json_client.h, td_log.h, ...
└── entrypoint.sh             # committed, see repo
```

## How to populate

Build TDLib once (any machine matching the image's glibc — the Dockerfile
uses `ubuntu:24.04`, glibc 2.39):

```sh
git clone https://github.com/tdlib/td.git
mkdir td/build && cd td/build
cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX:PATH=/usr/local ..
cmake --build . --target install
```

Then stage from the install prefix:

```sh
cp -L /usr/local/lib/libtdjson.so docker-libs/lib/libtdjson.so.1.8.64
ln -sf libtdjson.so.1.8.64 docker-libs/lib/libtdjson.so
cp -r /usr/local/include/td docker-libs/include/
```

Notes:
- The schema vendored in `v2/internal/gen/scheme/td_api.tl` is pinned to the
  TDLib version this library was built from. If you upgrade libtdjson,
  update the schema to match and regenerate (`cd v2 && go generate ./td`).
- The image only needs the shared library at runtime; the static archives
  (`libtdjson_static.a`, `libtdjson_private.a`) are not required.
