#!/usr/bin/env bash
# Guard: internal/<domain>/ tidak boleh import internal/<domain-lain>/
# (kecuali platform/ — itu bottom layer untuk semua).
#
# Setiap domain harus berkomunikasi via composition root di cmd/server/main.go,
# atau via interface yang diterima lewat platform.Deps.
set -euo pipefail

INTERNAL_DIR="internal"
[ -d "$INTERNAL_DIR" ] || exit 0

violations=0
for domain_dir in "$INTERNAL_DIR"/*/; do
    domain=$(basename "$domain_dir")
    [ "$domain" = "platform" ] && continue
    [ -d "$domain_dir" ] || continue

    # Cari import "project/internal/<other>" di file Go domain ini
    while IFS= read -r file; do
        bad=$(grep -E '"project/internal/[a-z]+(/[a-z]+)?"' "$file" \
            | grep -v '"project/internal/platform' \
            | grep -v "\"project/internal/${domain}\"" \
            | grep -v "\"project/internal/${domain}/" || true)
        if [ -n "$bad" ]; then
            echo "ERROR: $file imports another domain:"
            echo "$bad"
            violations=$((violations + 1))
        fi
    done < <(find "$domain_dir" -name '*.go' -type f)
done

if [ "$violations" -gt 0 ]; then
    echo ""
    echo "Cross-domain imports detected. Use platform.Deps to pass shared interfaces"
    echo "(e.g. middleware) from main.go instead of importing the producing domain."
    exit 1
fi
