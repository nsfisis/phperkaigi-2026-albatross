FROM emscripten/emsdk:5.0.1 AS wasm-builder

RUN git clone --depth=1 --branch=php-8.5.3 https://github.com/php/php-src

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        autoconf \
        bison \
        pkg-config \
        re2c \
        && \
    :

# Define ZEND_MM_ERROR=0 to suppress munmap() errors.
RUN cd php-src && \
    ./buildconf --force && \
    emconfigure ./configure \
        --disable-all \
        --disable-cgi \
        --disable-cli \
        --disable-fiber-asm \
        --disable-mbregex \
        --disable-opcache-jit \
        --disable-phpdbg \
        --enable-embed=static \
        --enable-mbstring \
        --enable-short-tags \
        --without-iconv \
        --without-libxml \
        --without-pcre-jit \
        --without-pdo-sqlite \
        --without-sqlite3 \
        && \
    EMCC_CFLAGS='-D ZEND_MM_ERROR=0' emmake make -j$(nproc) && \
    mv libs/libphp.a .. && \
    make clean && \
    git clean -fd && \
    :

COPY worker/php/php-wasm.c /src/

RUN cd php-src && \
    emcc \
        -c \
        -o php-wasm.o \
        -I . \
        -I TSRM \
        -I Zend \
        -I main \
        ../php-wasm.c \
        && \
    mv php-wasm.o .. && \
    make clean && \
    git clean -fd && \
    :

RUN emcc \
    -s ENVIRONMENT=node \
    -s ERROR_ON_UNDEFINED_SYMBOLS=0 \
    -s EXPORTED_RUNTIME_METHODS='["ccall"]' \
    -s EXPORT_ES6=1 \
    -s INITIAL_MEMORY=16777216 \
    -s INVOKE_RUN=0 \
    -s MODULARIZE=1 \
    -o php-wasm.js \
    php-wasm.o \
    libphp.a \
    ;

FROM node:22.14

WORKDIR /app
ENV NODE_ENV=production

COPY --from=wasm-builder /src/php-wasm.js /src/php-wasm.wasm /app/worker/php/
COPY package.json package-lock.json ./
COPY worker/php/package.json worker/php/
RUN npm install -w worker/php --omit=dev

COPY worker/php/index.mjs worker/php/lib.mjs worker/php/exec.mjs worker/php/

WORKDIR /app/worker/php
ENTRYPOINT ["node", "index.mjs"]
EXPOSE 80
