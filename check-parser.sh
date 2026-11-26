#!/bin/sh
# Checks that the CX parsers are what make build-parser writes from the .y
# grammars, so a rename in a parser is made in its grammar as well.
P="t5-cxpartialparsing/cxpartialparsing.go t8-cxparsingcompletor/parsingcompletor.go"
# goyacc on windows writes a few lines of the parsers differently, so the
# parsers are checked on Linux and macOS.
if [ "$(go env GOOS)" = windows ]; then
	echo "parsers are not checked on windows"
	exit 0
fi
git diff --quiet -- $P || { echo "commit or restore $P first"; exit 1; }
make -s build-parser >/dev/null || exit 1
if git diff --quiet -- $P; then
	echo "parsers match their grammars"
	exit 0
fi
git diff --stat -- $P
git checkout -- $P
echo "a parser differs from what its .y grammar generates, change the grammar and run make build-parser"
exit 1
