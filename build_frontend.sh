#!/bin/bash
set -e

ANGULAR_DIR="app"
BUILD_OUTPUT_DIR="$ANGULAR_DIR/dist"
TARGET_DIR="ui"

echo "[+] Building Angular..."
cd $ANGULAR_DIR
pnpm install
pnpm build --configuration production
cd ..

# Descobre a pasta de build real (ela pode variar: dist/<project-name>/browser ou dist/<project-name>)
BUILD_SUBDIR=$(find "$BUILD_OUTPUT_DIR" -maxdepth 1 -mindepth 1 -type d | head -n 1)

if [ ! -d "$BUILD_SUBDIR" ]; then
  echo "[✗] Build output not found in: $BUILD_SUBDIR"
  exit 1
fi

echo "[+] Copying dist contents from $BUILD_SUBDIR to $TARGET_DIR..."
rm -rf "$TARGET_DIR"
mkdir -p "$TARGET_DIR"
cp -r "$BUILD_SUBDIR"/browser/* "$TARGET_DIR"

echo "[✓] Frontend build complete."
echo "Output location: $(realpath $TARGET_DIR)"