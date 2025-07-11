# export NDK_HOME=/opt/android-ndk-r27c
# export CC_PATH=$NDK_HOME/toolchains/llvm/prebuilt/linux-x86_64/bin

# CGO_ENABLED=1 GOOS=android GOARCH=arm CC=$CC_PATH/armv7a-linux-androideabi35-clang go build -buildmode=c-shared -o ./lib/armeabi-v7a/ ./client/golib
# CGO_ENABLED=1 GOOS=android GOARCH=arm64 CC=$CC_PATH/aarch64-linux-android35-clang go build -buildmode=c-shared -o ./lib/arm64-v8a/ ./client/golib
# CGO_ENABLED=1 GOOS=android GOARCH=amd64 CC=$CC_PATH/x86_64-linux-android35-clang go build -buildmode=c-shared -o ./lib/x86-64/ ./client/golib


# out NDK home path
NDK_PATH=/opt/android-ndk-r27c

CC_PATH=$NDK_PATH/toolchains/llvm/prebuilt/linux-x86_64/bin
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

OUT_PATH=$SCRIPT_DIR/output

# our lib name
TARGET=libgoandroid.so

LDFLAGS='-extldflags -Wl,-soname,'"$TARGET"''

GOOS=android GOARCH=amd64 CGO_ENABLED=1 CC=$CC_PATH/x86_64-linux-android34-clang go build -ldflags ''"$LDFLAGS"'' -buildmode=c-shared -o $OUT_PATH/x86_64/$TARGET ./client/golib

GOOS=android GOARCH=386 CGO_ENABLED=1 CC=$CC_PATH/i686-linux-android34-clang go build -ldflags ''"$LDFLAGS"'' -buildmode=c-shared -o $OUT_PATH/x86/$TARGET ./client/golib

GOOS=android GOARCH=arm64 CGO_ENABLED=1 CC=$CC_PATH/aarch64-linux-android34-clang go build -ldflags ''"$LDFLAGS"'' -buildmode=c-shared -o $OUT_PATH/arm64-v8a/$TARGET ./client/golib

GOOS=android GOARCH=arm CGO_ENABLED=1 CC=$CC_PATH/armv7a-linux-androideabi34-clang go build -ldflags ''"$LDFLAGS"'' -buildmode=c-shared -o $OUT_PATH/armeabi-v7a/$TARGET ./client/golib
