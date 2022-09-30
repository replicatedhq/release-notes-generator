#!/bin/sh

set -e

eval set -- "$@"

RELEASE_NOTES="$(/release-notes-generator "$@")"

RELEASE_NOTES="${RELEASE_NOTES//'%'/'%25'}"
RELEASE_NOTES="${RELEASE_NOTES//$'\n'/'%0A'}"
RELEASE_NOTES="${RELEASE_NOTES//$'\r'/'%0D'}"

echo "::set-output name=release-notes::$RELEASE_NOTES"
