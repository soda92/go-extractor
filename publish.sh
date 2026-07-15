#!/bin/bash
set -e

# Configuration
PKGNAME="go-extractor"
REMOTE_HOST="ubuntu"
REMOTE_DIR="/var/www/pacman"

echo "==> Cleaning old local packages..."
rm -f ${PKGNAME}-[0-9]*-x86_64.pkg.tar.zst ${PKGNAME}-debug-*-x86_64.pkg.tar.zst

echo "==> Building package with makepkg..."
makepkg -cf --noconfirm

# Find the generated package file
PKG_FILE=$(ls -1t ${PKGNAME}-[0-9]*-x86_64.pkg.tar.zst | head -n 1)

if [ -z "$PKG_FILE" ]; then
    echo "ERROR: Package file not found!"
    exit 1
fi

echo "==> Copying package to local repository structure..."
mkdir -p repo/x86_64
cp "$PKG_FILE" repo/x86_64/

echo "==> Updating database with repo-add..."
repo-add repo/x86_64/custom.db.tar.zst repo/x86_64/"$PKG_FILE"

echo "==> Preparing remote directory..."
ssh "$REMOTE_HOST" "sudo mkdir -p $REMOTE_DIR && sudo chown -R \$(whoami) $REMOTE_DIR"

echo "==> Syncing files via rsync..."
rsync -avz --delete repo/x86_64/ "$REMOTE_HOST":"$REMOTE_DIR/x86_64/"

echo "==> Cleaning up build artifacts..."
rm -f "$PKG_FILE"

echo "==> Success! Remote repository has been updated."
