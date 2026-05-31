#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
UPDATE_SCRIPT="${SCRIPT_DIR}/update_production.sh"

TMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/update-production-test-XXXXXX")"
trap 'rm -rf "${TMP_ROOT}"' EXIT

REPO_ROOT="${TMP_ROOT}/repo"
RELEASE_DIR="${TMP_ROOT}/release"
BACKUP_DIR="${TMP_ROOT}/backup"

mkdir -p "${REPO_ROOT}/deploy/ops" "${RELEASE_DIR}" "${BACKUP_DIR}"
cp "${UPDATE_SCRIPT}" "${REPO_ROOT}/deploy/ops/update_production.sh"
SOURCEABLE_SCRIPT="${TMP_ROOT}/update_production.source.sh"
sed '$d' "${UPDATE_SCRIPT}" > "${SOURCEABLE_SCRIPT}"

(
  cd "${REPO_ROOT}"
  git init -q
  git config user.email "test@example.com"
  git config user.name "Update Script Test"

  mkdir -p backend
  printf 'tracked from head\n' > backend/tracked.txt
  git add backend/tracked.txt
  git commit -qm "init"

  printf 'tracked local override\n' > backend/tracked.txt
  mkdir -p backend/new
  printf 'untracked local file\n' > backend/new/untracked.txt

  dirty_tracked_paths="$(git ls-files -m -d || true)"
  untracked_paths="$(git ls-files --others --exclude-standard || true)"
  target_full_sha="$(git rev-parse HEAD)"
  release_dir="${RELEASE_DIR}"
  dirty_tracked_list_file="${BACKUP_DIR}/dirty-tracked-files.txt"
  dirty_tracked_tar="${BACKUP_DIR}/dirty-tracked-files.tar"
  untracked_list_file="${BACKUP_DIR}/untracked-files.txt"
  untracked_tar="${BACKUP_DIR}/untracked-files.tar"

  printf '%s\n' "${dirty_tracked_paths}" > "${dirty_tracked_list_file}"
  tar -C "${REPO_ROOT}" -cf "${dirty_tracked_tar}" -T "${dirty_tracked_list_file}"

  printf '%s\n' "${untracked_paths}" > "${untracked_list_file}"
  tar -C "${REPO_ROOT}" -cf "${untracked_tar}" -T "${untracked_list_file}"

  if UPDATE_SCRIPT_PATH="${SOURCEABLE_SCRIPT}" bash -lc '
    set -euo pipefail
    source "${UPDATE_SCRIPT_PATH}"
    trap - EXIT
    ensure_repo_state
  ' > "${TMP_ROOT}/default-run.log" 2>&1; then
    printf 'expected ensure_repo_state to fail for dirty worktree by default\n' >&2
    exit 1
  fi

  grep -q "Refusing to build from a dirty worktree" "${TMP_ROOT}/default-run.log"

  UPDATE_SCRIPT_PATH="${SOURCEABLE_SCRIPT}" ALLOW_DIRTY_OVERLAY=1 bash -lc '
    set -euo pipefail
    source "${UPDATE_SCRIPT_PATH}"
    trap - EXIT
    ensure_repo_state
    RELEASE_DIR="'"${release_dir}"'"
    TARGET_FULL_SHA="'"${target_full_sha}"'"
    DIRTY_TRACKED_TAR="'"${dirty_tracked_tar}"'"
    UNTRACKED_TAR="'"${untracked_tar}"'"
    export RELEASE_DIR TARGET_FULL_SHA DIRTY_TRACKED_TAR UNTRACKED_TAR
    git archive --format=tar "${TARGET_FULL_SHA}" | tar -xf - -C "${RELEASE_DIR}"
    overlay_preserved_files_into_release_dir
  '

  test "$(cat "${RELEASE_DIR}/backend/tracked.txt")" = "tracked local override"
  test "$(cat "${RELEASE_DIR}/backend/new/untracked.txt")" = "untracked local file"
)

OVERRIDE_FILE="${TMP_ROOT}/docker-compose.override.yml"
cat > "${OVERRIDE_FILE}" <<'YAML'
services:
  sub2api:
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - KEEP_CUSTOM=1
  redis:
    image: redis:7
YAML

UPDATE_SCRIPT_PATH="${SOURCEABLE_SCRIPT}" OVERRIDE_FILE="${OVERRIDE_FILE}" bash -lc '
  set -euo pipefail
  source "${UPDATE_SCRIPT_PATH}"
  trap - EXIT
  OVERRIDE_FILE="'"${OVERRIDE_FILE}"'"
  SERVICE_NAME=sub2api
  IMAGE_REPOSITORY=openapi-prod
  TARGET_SHORT_SHA=abc12345
  write_override_file
'

grep -q "image: openapi-prod:abc12345" "${OVERRIDE_FILE}"
grep -q "KEEP_CUSTOM=1" "${OVERRIDE_FILE}"
grep -q "redis:" "${OVERRIDE_FILE}"
grep -q "image: redis:7" "${OVERRIDE_FILE}"

FAKE_BIN="${TMP_ROOT}/bin"
mkdir -p "${FAKE_BIN}" "${TMP_ROOT}/deploy"
cat > "${FAKE_BIN}/docker" <<'SH'
#!/usr/bin/env bash
printf '%s\n' "$*" > "${DOCKER_ARGS_FILE}"
SH
chmod +x "${FAKE_BIN}/docker"

UPDATE_SCRIPT_PATH="${SOURCEABLE_SCRIPT}" DOCKER_ARGS_FILE="${TMP_ROOT}/docker.args" PATH="${FAKE_BIN}:${PATH}" bash -lc '
  set -euo pipefail
  source "${UPDATE_SCRIPT_PATH}"
  trap - EXIT
  DEPLOY_DIR="'"${TMP_ROOT}/deploy"'"
  SERVICE_NAME=sub2api
  deploy_service
'

grep -q -- "compose up -d --no-deps sub2api" "${TMP_ROOT}/docker.args"

printf 'update_production.sh preserves override customizations, blocks dirty worktrees, and restarts only the app service\n'
