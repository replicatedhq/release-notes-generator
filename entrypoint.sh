#!/bin/sh

CMD=$1

if [ -z "$CMD" ]; then
    echo "No command supplied"
fi

RELEASE_NOTES=$(eval $CMD)

echo "::set-output name=release-notes::$RELEASE_NOTES"