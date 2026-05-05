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

  # shellcheck disable=SC1090
  source <(sed '$d' "${UPDATE_SCRIPT}")
  trap - EXIT

  RELEASE_DIR="${release_dir}"
  TARGET_FULL_SHA="${target_full_sha}"
  DIRTY_TRACKED_TAR="${dirty_tracked_tar}"
  UNTRACKED_TAR="${untracked_tar}"
  export RELEASE_DIR TARGET_FULL_SHA DIRTY_TRACKED_TAR UNTRACKED_TAR
  git archive --format=tar "${TARGET_FULL_SHA}" | tar -xf - -C "${RELEASE_DIR}"
  overlay_preserved_files_into_release_dir

  test "$(cat "${RELEASE_DIR}/backend/tracked.txt")" = "tracked local override"
  test "$(cat "${RELEASE_DIR}/backend/new/untracked.txt")" = "untracked local file"
)

printf 'update_production.sh preserves tracked and untracked local files in release builds\n'
