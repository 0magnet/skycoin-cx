#!/bin/sh
# Checks the init rule in README.md. No func init() outside package main,
# except the files listed in init-allowlist.txt, which must only shrink.
A=init-allowlist.txt
F=$(mktemp)
find . -path ./vendor -prune -o -path ./.git -prune -o -path ./.claude -prune -o -name '*.go' -print |
	xargs grep -l '^func init()' | sed 's#^\./##' | while read -r f; do
	grep -q '^package main$' "$f" || echo "$f"
done | sort >"$F"
rc=0
for f in $(comm -23 "$F" "$A"); do echo "func init() outside main, call an explicit Init from main instead: $f"; rc=1; done
for f in $(comm -13 "$F" "$A"); do echo "fixed, remove from $A: $f"; rc=1; done
echo "$(wc -l <"$F") allowlisted init() files"
rm -f "$F"
exit $rc
