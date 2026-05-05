#!/bin/sh
set -e

# Fix data directory permissions when running as root.
# Docker named volumes / host bind-mounts may be owned by root,
# preventing the non-root sub2api user from writing files.
if [ "$(id -u)" = "0" ]; then
    mkdir -p /app/data
    # Use || true to avoid failure on read-only mounted files (e.g. config.yaml:ro)
    chown -R sub2api:sub2api /app/data 2>/dev/null || true
    if [ -S /var/run/docker.sock ]; then
        DOCKER_GID="$(stat -c '%g' /var/run/docker.sock 2>/dev/null || true)"
        if [ -n "${DOCKER_GID}" ]; then
            if ! awk -F: -v gid="${DOCKER_GID}" '$3 == gid { found = 1 } END { exit !found }' /etc/group; then
                addgroup -g "${DOCKER_GID}" dockersock 2>/dev/null || true
            fi
            DOCKER_GROUP="$(awk -F: -v gid="${DOCKER_GID}" '$3 == gid { print $1; exit }' /etc/group)"
            if [ -n "${DOCKER_GROUP}" ]; then
                addgroup sub2api "${DOCKER_GROUP}" 2>/dev/null || true
            fi
        fi
    fi
    # Re-invoke this script as sub2api so the flag-detection below
    # also runs under the correct user.
    exec su-exec sub2api "$0" "$@"
fi

# Compatibility: if the first arg looks like a flag (e.g. --help),
# prepend the default binary so it behaves the same as the old
# ENTRYPOINT ["/app/sub2api"] style.
if [ "${1#-}" != "$1" ]; then
    set -- /app/sub2api "$@"
fi

exec "$@"
