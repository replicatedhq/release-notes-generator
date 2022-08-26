#!/bin/sh

CMD=$1

if [ -z "$CMD" ]; then
    echo "No command supplied"
fi

RELEASE_NOTES=$(eval $CMD)

echo "RELEASE_NOTES<<EOF" >> $GITHUB_ENV
echo "$RELEASE_NOTES" >> $GITHUB_ENV
echo "EOF" >> $GITHUB_ENV

echo "::set-output name=release-notes::$RELEASE_NOTES"