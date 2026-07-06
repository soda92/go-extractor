# Maintainer: Maple <wjxa20152015@gmail.com>
pkgname=go-extractor
pkgver=0.1.0
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
}

build() {
  export GOPATH="$srcdir/gopath"
  export GOCACHE="$srcdir/gocache"
  
  # Standard Go build options for Arch packaging
  export CGO_ENABLED=1
  export CGO_LDFLAGS="${LDFLAGS}"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  
  go build \
    -buildmode=pie \
    -trimpath \
    -ldflags="-linkmode=external" \
    -mod=readonly \
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
