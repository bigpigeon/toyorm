diff=$(gofmt -s -d .);if [ -n "$diff" ]; then exit 1; fi;
