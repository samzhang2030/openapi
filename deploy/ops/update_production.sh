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

cleanup() {
  if [ -n "${RELEASE_DIR}" ] && [ -d "${RELEASE_DIR}" ]; then
    rm -rf "${RELEASE_DIR}"
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
  local dirty_tracked
  dirty_tracked="$(git status --porcelain --untracked-files=no || true)"
  dirty_tracked="$(printf '%s\n' "${dirty_tracked}" | grep -vE '^[ MADRCU?]{2} deploy/docker-compose\.yml$' | grep -vE '^[ MADRCU?]{2} deploy/docker-compose\.override\.yml$' || true)"

  if [ -n "${dirty_tracked}" ]; then
    print_error "Tracked local changes outside deploy compose files would make update unsafe:"
    printf '%s\n' "${dirty_tracked}" >&2
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

  if [ -f "${LOCAL_COMPOSE_FILE}" ]; then
    cp -a "${LOCAL_COMPOSE_FILE}" "${BACKUP_DIR}/docker-compose.yml"
  fi

  if [ -f "${OVERRIDE_FILE}" ]; then
    cp -a "${OVERRIDE_FILE}" "${BACKUP_DIR}/docker-compose.override.yml"
  fi
}

restore_local_compose() {
  if [ -f "${BACKUP_DIR}/docker-compose.yml" ]; then
    cp -a "${BACKUP_DIR}/docker-compose.yml" "${LOCAL_COMPOSE_FILE}"
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
  git restore --worktree --staged -- deploy/docker-compose.yml 2>/dev/null || true
  git merge --ff-only "${TARGET_FULL_SHA}"
  restore_local_compose
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

  print_info "Building image ${image_ref}..."
  docker build \
    --pull \
    --build-arg NODE_MAX_OLD_SPACE_SIZE="${NODE_MAX_OLD_SPACE_SIZE}" \
    -t "${image_ref}" \
    "${RELEASE_DIR}"

  print_success "Built image ${image_ref}"
}

write_override_file() {
  mkdir -p "$(dirname "${OVERRIDE_FILE}")"
  cat > "${OVERRIDE_FILE}" <<EOF
services:
  ${SERVICE_NAME}:
    image: ${IMAGE_REPOSITORY}:${TARGET_SHORT_SHA}
EOF
  print_success "Updated ${OVERRIDE_FILE}"
}

deploy_service() {
  print_info "Restarting ${SERVICE_NAME} with docker compose..."
  (
    cd "${DEPLOY_DIR}"
    docker compose up -d "${SERVICE_NAME}"
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
