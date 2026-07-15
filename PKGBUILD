# Maintainer: Maple <wjxa20152015@gmail.com>
pkgname=go-extractor
pkgver=0.2.0
pkgrel=1
pkgdesc="A Fyne-based GUI tool for extracting archives to a custom subfolder, integrated with Dolphin."
arch=('x86_64')
url="https://github.com/wjxa2015/go-extractor"
license=('unknown')
depends=('7zip' 'glibc' 'libgl' 'libx11' 'libxrandr' 'libxxf86vm' 'libxi' 'libxcursor' 'libxinerama')
makedepends=('go')
source=('main.go'
        'go.mod'
        'go.sum'
        'go-extractor.desktop')
sha256sums=('SKIP'
            'SKIP'
            'SKIP'
            'SKIP')

prepare() {
  # Create build directory
  mkdir -p "$srcdir/build"

  # Copy vendor folder if it exists in the build start directory to speed up building offline
  if [ -d "$startdir/vendor" ]; then
    cp -r "$startdir/vendor" "$srcdir/vendor"
  fi
}

build() {
  # By default, we comment these out to leverage your user-level Go build cache (~/.cache/go-build),
  # which speeds up compiles (including CGO/Fyne compiling) to 1-2 seconds.
  # If doing a strict, clean chroot build, uncomment these to isolate the environment.
  # export GOPATH="$srcdir/gopath"
  # export GOCACHE="$srcdir/gocache"
  
  # Speed up CGO compilation using ccache if available
  if command -v ccache >/dev/null 2>&1; then
    export CC="ccache gcc"
    export CXX="ccache g++"
    export CCACHE_NOHASHDIR=1
    export CCACHE_BASEDIR="$srcdir"
  fi

  # Standard Go build options for Arch packaging
  export CGO_ENABLED=1
  export CGO_LDFLAGS="${LDFLAGS}"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  
  # Use -mod=vendor if vendor directory exists
  local GO_MOD_FLAG="-mod=readonly"
  if [ -d "$srcdir/vendor" ]; then
    GO_MOD_FLAG="-mod=vendor"
  fi

  go build \
    -buildmode=pie \
    -trimpath \
    -ldflags="-linkmode=external" \
    $GO_MOD_FLAG \
    -modcacherw \
    -o "$srcdir/build/go-extractor" \
    "$srcdir/main.go"
}

package() {
  # Install the binary
  install -Dm755 "$srcdir/build/go-extractor" "$pkgdir/usr/bin/go-extractor"

  # Install the Dolphin service menu
  install -Dm644 "$srcdir/go-extractor.desktop" "$pkgdir/usr/share/kio/servicemenus/go-extractor.desktop"
}
