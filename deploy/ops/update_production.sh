#!/usr/bin/env bash

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
  printf '%b\n' "${BLUE}[INFO]${NC} $1"
}

print_success() {
  printf '%b\n' "${GREEN}[OK]${NC} $1"
}

print_warning() {
  printf '%b\n' "${YELLOW}[WARN]${NC} $1"
}

print_error() {
  printf '%b\n' "${RED}[ERROR]${NC} $1" >&2
}

require_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    print_error "Missing required command: $1"
    exit 1
  fi
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEPLOY_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
REPO_ROOT="$(cd "${DEPLOY_DIR}/.." && pwd)"

GIT_REMOTE="${GIT_REMOTE:-origin}"
TARGET_REF="${1:-${GIT_REMOTE}/main}"
SERVICE_NAME="${SERVICE_NAME:-sub2api}"
IMAGE_REPOSITORY="${IMAGE_REPOSITORY:-openapi-prod}"
NODE_MAX_OLD_SPACE_SIZE="${NODE_MAX_OLD_SPACE_SIZE:-4096}"
FORCE_BUILD="${FORCE_BUILD:-0}"
ALLOW_DIRTY_OVERLAY="${ALLOW_DIRTY_OVERLAY:-0}"
LOCAL_HEALTH_URL="${LOCAL_HEALTH_URL:-http://127.0.0.1:8080/health}"
PUBLIC_HEALTH_URL="${PUBLIC_HEALTH_URL:-}"
OVERRIDE_FILE="${DEPLOY_DIR}/docker-compose.override.yml"
LOCAL_COMPOSE_FILE="${DEPLOY_DIR}/docker-compose.yml"
BACKUP_ROOT="${DEPLOY_DIR}/.backups"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
BACKUP_DIR="${BACKUP_ROOT}/update-${TIMESTAMP}"
RELEASE_DIR=""
TARGET_FULL_SHA=""
TARGET_SHORT_SHA=""
DIRTY_TRACKED_PATHS=""
DIRTY_TRACKED_LIST_FILE=""
DIRTY_TRACKED_TAR=""
UNTRACKED_PATHS=""
UNTRACKED_LIST_FILE=""
UNTRACKED_TAR=""
WORKTREE_CLEANED_FOR_MERGE=0
PRESERVED_FILES_RESTORED=0

cleanup() {
  if [ -n "${RELEASE_DIR}" ] && [ -d "${RELEASE_DIR}" ]; then
    rm -rf "${RELEASE_DIR}"
  fi
  if [ "${WORKTREE_CLEANED_FOR_MERGE}" = "1" ] && [ "${PRESERVED_FILES_RESTORED}" = "0" ]; then
    print_warning "Restoring preserved local files after interrupted update..."
    restore_local_files || true
  fi
}

trap cleanup EXIT

http_get() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsS "$1"
  elif command -v wget >/dev/null 2>&1; then
    wget -qO- "$1"
  else
    print_error "Neither curl nor wget is available for health checks"
    exit 1
  fi
}

ensure_repo_state() {
  DIRTY_TRACKED_PATHS="$(git ls-files -m -d || true)"
  if [ -n "${DIRTY_TRACKED_PATHS}" ] && [ "${ALLOW_DIRTY_OVERLAY}" = "1" ]; then
    print_warning "Preserving tracked local changes during update:"
    printf '%s\n' "${DIRTY_TRACKED_PATHS}"
  fi

  UNTRACKED_PATHS="$(git ls-files --others --exclude-standard || true)"
  if [ -n "${UNTRACKED_PATHS}" ] && [ "${ALLOW_DIRTY_OVERLAY}" = "1" ]; then
    print_warning "Including untracked local files in build context:"
    printf '%s\n' "${UNTRACKED_PATHS}"
  fi

  if [ "${ALLOW_DIRTY_OVERLAY}" != "1" ] && { [ -n "${DIRTY_TRACKED_PATHS}" ] || [ -n "${UNTRACKED_PATHS}" ]; }; then
    print_error "Refusing to build from a dirty worktree. Commit/stash changes first, or set ALLOW_DIRTY_OVERLAY=1 for an explicit override."
    if [ -n "${DIRTY_TRACKED_PATHS}" ]; then
      print_error "Tracked changes:"
      printf '%s\n' "${DIRTY_TRACKED_PATHS}" >&2
    fi
    if [ -n "${UNTRACKED_PATHS}" ]; then
      print_error "Untracked files:"
      printf '%s\n' "${UNTRACKED_PATHS}" >&2
    fi
    exit 1
  fi
}

check_untracked_conflicts() {
  local untracked conflict_paths path

  untracked="$(git ls-files --others --exclude-standard || true)"
  conflict_paths=""

  while IFS= read -r path; do
    [ -n "${path}" ] || continue
    if git cat-file -e "${TARGET_FULL_SHA}:${path}" 2>/dev/null; then
      conflict_paths="${conflict_paths}${path}"$'\n'
    fi
  done <<< "${untracked}"

  if [ -n "${conflict_paths}" ]; then
    print_error "Untracked files would be overwritten by ${TARGET_REF}:"
    printf '%s' "${conflict_paths}" >&2
    exit 1
  fi
}

backup_local_files() {
  mkdir -p "${BACKUP_DIR}"
  DIRTY_TRACKED_LIST_FILE="${BACKUP_DIR}/dirty-tracked-files.txt"
  DIRTY_TRACKED_TAR="${BACKUP_DIR}/dirty-tracked-files.tar"

  if [ -f "${LOCAL_COMPOSE_FILE}" ]; then
    cp -a "${LOCAL_COMPOSE_FILE}" "${BACKUP_DIR}/docker-compose.yml"
  fi

  if [ -f "${OVERRIDE_FILE}" ]; then
    cp -a "${OVERRIDE_FILE}" "${BACKUP_DIR}/docker-compose.override.yml"
  fi

  if [ -n "${DIRTY_TRACKED_PATHS}" ]; then
    printf '%s\n' "${DIRTY_TRACKED_PATHS}" > "${DIRTY_TRACKED_LIST_FILE}"
    tar -C "${REPO_ROOT}" -cf "${DIRTY_TRACKED_TAR}" -T "${DIRTY_TRACKED_LIST_FILE}"
  fi

  if [ -n "${UNTRACKED_PATHS}" ]; then
    UNTRACKED_LIST_FILE="${BACKUP_DIR}/untracked-files.txt"
    UNTRACKED_TAR="${BACKUP_DIR}/untracked-files.tar"
    printf '%s\n' "${UNTRACKED_PATHS}" > "${UNTRACKED_LIST_FILE}"
    tar -C "${REPO_ROOT}" -cf "${UNTRACKED_TAR}" -T "${UNTRACKED_LIST_FILE}"
  fi
}

restore_local_files() {
  if [ -f "${BACKUP_DIR}/docker-compose.yml" ]; then
    cp -a "${BACKUP_DIR}/docker-compose.yml" "${LOCAL_COMPOSE_FILE}"
  fi

  if [ -f "${BACKUP_DIR}/docker-compose.override.yml" ]; then
    cp -a "${BACKUP_DIR}/docker-compose.override.yml" "${OVERRIDE_FILE}"
  fi

  if [ -f "${DIRTY_TRACKED_TAR}" ]; then
    tar -C "${REPO_ROOT}" -xf "${DIRTY_TRACKED_TAR}"
  fi
  PRESERVED_FILES_RESTORED=1
}

clean_worktree_for_merge() {
  if [ -n "${DIRTY_TRACKED_PATHS}" ]; then
    print_info "Temporarily resetting tracked local changes before merge..."
    git restore --worktree --staged .
    WORKTREE_CLEANED_FOR_MERGE=1
  fi
}

overlay_preserved_files_into_release_dir() {
  if [ -f "${DIRTY_TRACKED_TAR}" ]; then
    tar -C "${RELEASE_DIR}" -xf "${DIRTY_TRACKED_TAR}"
  fi
  if [ -f "${UNTRACKED_TAR}" ]; then
    tar -C "${RELEASE_DIR}" -xf "${UNTRACKED_TAR}"
  fi
}

update_repo() {
  print_info "Fetching ${GIT_REMOTE}..."
  git fetch "${GIT_REMOTE}" --prune

  TARGET_FULL_SHA="$(git rev-parse "${TARGET_REF}")"
  TARGET_SHORT_SHA="$(git rev-parse --short=8 "${TARGET_FULL_SHA}")"

  check_untracked_conflicts

  if ! git diff --quiet HEAD.."${TARGET_FULL_SHA}" -- deploy/docker-compose.yml; then
    print_warning "Target commit changes deploy/docker-compose.yml, but the server-local file will be preserved."
  fi

  local current_sha
  current_sha="$(git rev-parse HEAD)"

  if [ "${current_sha}" = "${TARGET_FULL_SHA}" ]; then
    print_info "Repository already at ${TARGET_SHORT_SHA}"
    return
  fi

  print_info "Fast-forwarding repository to ${TARGET_SHORT_SHA}..."
  clean_worktree_for_merge
  git merge --ff-only "${TARGET_FULL_SHA}"
  restore_local_files
  print_success "Repository updated to ${TARGET_SHORT_SHA}"
}

build_release_image() {
  local image_ref
  image_ref="${IMAGE_REPOSITORY}:${TARGET_SHORT_SHA}"

  if docker image inspect "${image_ref}" >/dev/null 2>&1 && [ "${FORCE_BUILD}" != "1" ]; then
    print_info "Image ${image_ref} already exists locally, skipping build"
    return
  fi

  RELEASE_DIR="$(mktemp -d "${TMPDIR:-/tmp}/openapi-release-${TARGET_SHORT_SHA}-XXXXXX")"
  print_info "Preparing clean build context at ${RELEASE_DIR}"
  git archive --format=tar "${TARGET_FULL_SHA}" | tar -xf - -C "${RELEASE_DIR}"
  overlay_preserved_files_into_release_dir

  print_info "Building image ${image_ref}..."
  docker build \
    --pull \
    --build-arg NODE_MAX_OLD_SPACE_SIZE="${NODE_MAX_OLD_SPACE_SIZE}" \
    -t "${image_ref}" \
    "${RELEASE_DIR}"

  print_success "Built image ${image_ref}"
}

write_override_file() {
  local image_ref
  image_ref="${IMAGE_REPOSITORY}:${TARGET_SHORT_SHA}"

  mkdir -p "$(dirname "${OVERRIDE_FILE}")"
  python3 - "${OVERRIDE_FILE}" "${SERVICE_NAME}" "${image_ref}" <<'PY'
import pathlib
import re
import sys

path = pathlib.Path(sys.argv[1])
service_name = sys.argv[2]
image_ref = sys.argv[3]

lines = path.read_text().splitlines(keepends=True) if path.exists() else []


def line_indent(line: str) -> int:
    return len(line) - len(line.lstrip(" "))


def is_block_boundary(line: str, indent: int) -> bool:
    stripped = line.strip()
    return bool(stripped and not stripped.startswith("#") and line_indent(line) <= indent)


def append_block() -> None:
    if lines and not lines[-1].endswith("\n"):
        lines[-1] = lines[-1] + "\n"
    if lines and lines[-1].strip():
        lines.append("\n")
    lines.extend([
        "services:\n",
        f"  {service_name}:\n",
        f"    image: {image_ref}\n",
    ])


services_idx = None
for idx, line in enumerate(lines):
    if re.match(r"^services\s*:\s*(?:#.*)?$", line):
        services_idx = idx
        break

if services_idx is None:
    append_block()
    path.write_text("".join(lines))
    sys.exit(0)

services_indent = line_indent(lines[services_idx])
services_end = len(lines)
for idx in range(services_idx + 1, len(lines)):
    if is_block_boundary(lines[idx], services_indent):
        services_end = idx
        break

service_idx = None
service_indent = services_indent + 2
service_pattern = re.compile(rf"^ {{{service_indent}}}{re.escape(service_name)}\s*:\s*(?:#.*)?$")
for idx in range(services_idx + 1, services_end):
    if service_pattern.match(lines[idx]):
        service_idx = idx
        break

if service_idx is None:
    insert_at = services_end
    lines[insert_at:insert_at] = [
        f"  {service_name}:\n",
        f"    image: {image_ref}\n",
    ]
    path.write_text("".join(lines))
    sys.exit(0)

service_end = services_end
for idx in range(service_idx + 1, services_end):
    if is_block_boundary(lines[idx], service_indent):
        service_end = idx
        break

image_pattern = re.compile(r"^(\s*)image\s*:")
for idx in range(service_idx + 1, service_end):
    match = image_pattern.match(lines[idx])
    if match:
        lines[idx] = f"{match.group(1)}image: {image_ref}\n"
        path.write_text("".join(lines))
        sys.exit(0)

lines[service_idx + 1:service_idx + 1] = [f"    image: {image_ref}\n"]
path.write_text("".join(lines))
PY
  print_success "Updated ${OVERRIDE_FILE} service image to ${image_ref}"
}

deploy_service() {
  print_info "Restarting ${SERVICE_NAME} with docker compose..."
  (
    cd "${DEPLOY_DIR}"
    docker compose up -d --no-deps "${SERVICE_NAME}"
  )
}

wait_for_service_health() {
  local attempt max_attempts status image_ref
  max_attempts=60

  for attempt in $(seq 1 "${max_attempts}"); do
    status="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "${SERVICE_NAME}" 2>/dev/null || true)"
    if [ "${status}" = "healthy" ] || [ "${status}" = "running" ]; then
      image_ref="$(docker inspect -f '{{.Config.Image}}' "${SERVICE_NAME}")"
      print_success "${SERVICE_NAME} is ${status} on image ${image_ref}"
      return
    fi
    sleep 2
  done

  print_error "${SERVICE_NAME} did not become healthy in time"
  docker ps -a --filter "name=${SERVICE_NAME}"
  docker logs --tail 120 "${SERVICE_NAME}" || true
  exit 1
}

run_health_checks() {
  print_info "Checking local health endpoint..."
  http_get "${LOCAL_HEALTH_URL}" >/dev/null
  print_success "Local health endpoint responded"

  if [ -n "${PUBLIC_HEALTH_URL}" ]; then
    print_info "Checking public health endpoint..."
    http_get "${PUBLIC_HEALTH_URL}" >/dev/null
    print_success "Public health endpoint responded"
  fi
}

main() {
  require_command git
  require_command docker
  require_command python3
  require_command tar
  require_command mktemp

  if ! docker compose version >/dev/null 2>&1; then
    print_error "docker compose is required"
    exit 1
  fi

  cd "${REPO_ROOT}"

  if [ "$(git rev-parse --show-toplevel)" != "${REPO_ROOT}" ]; then
    print_error "Could not resolve repository root"
    exit 1
  fi

  print_info "Repository root: ${REPO_ROOT}"
  ensure_repo_state
  backup_local_files
  update_repo
  build_release_image
  write_override_file
  deploy_service
  wait_for_service_health
  run_health_checks

  print_success "Production update finished at commit ${TARGET_SHORT_SHA}"
}

main "$@"
