#!/bin/sh
set -eu
root=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
tools="$root/tools"
assets="$root/app/web/assets"
editor="$root/admin/node_modules/@dawnim/admin/dist"
if [ ! -f "$editor/dawn-editor.js" ]; then
    echo '缺少编辑器产物：先在 admin 执行 pnpm install 安装 @dawnim/admin。' >&2
    exit 1
fi
markdown=$(cd "$tools" && node -p "require.resolve('markdown-it')")
"$tools/node_modules/.bin/esbuild" "$tools/moment_bundle.mjs" \
    --bundle --format=iife --platform=browser --target=es2022 \
    --alias:markdown-it="$markdown" \
    --outfile="$assets/.moment-render.js" --legal-comments=none
cat "$editor/dawn-editor.js" "$assets/.moment-render.js" > "$assets/moment-markdown.js"
rm "$assets/.moment-render.js"
cp "$editor/dawn-editor.css" "$assets/moment-markdown.css"
