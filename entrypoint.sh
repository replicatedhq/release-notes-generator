#!/bin/sh

set -e

eval set -- "$@"

RELEASE_NOTES="$(/release-notes-generator "$@")"

# GitHub disabled the legacy `::set-output` workflow command, so write the
# (multi-line) result to the step output via the environment file instead.
# A random delimiter avoids any collision with the release-notes content.
delimiter="ghadelimiter_$(head -c 16 /dev/urandom | od -An -tx1 | tr -d ' \n')"
{
  printf 'release-notes<<%s\n' "$delimiter"
  printf '%s\n' "$RELEASE_NOTES"
  printf '%s\n' "$delimiter"
} >> "$GITHUB_OUTPUT"
