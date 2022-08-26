#!/bin/sh

CMD=$1

if [ -z "$CMD" ]; then
    echo "No command supplied"
fi

RELEASE_NOTES=$(eval $CMD)

RELEASE_NOTES="${RELEASE_NOTES//'%'/'%25'}"
RELEASE_NOTES="${RELEASE_NOTES//$'\n'/'%0A'}"
RELEASE_NOTES="${RELEASE_NOTES//$'\r'/'%0D'}"

echo "::set-output name=release-notes::$RELEASE_NOTES" 