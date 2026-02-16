#!/usr/bin/env bash
set -Eeuo pipefail
set -o errtrace

LOG_FILE="/var/log/lightningos-app-upgrade.log"
mkdir -p /var/log /var/lib/lightningos /opt/lightningos/manager /opt/lightningos/ui
exec > >(tee -a "$LOG_FILE") 2>&1

print_step() {
  echo ""
  echo "==> $1"
}

print_ok() {
  echo "[OK] $1"
}

print_warn() {
  echo "[WARN] $1"
}

die() {
  echo "[ERROR] $1" >&2
  exit 1
}

require_root() {
  if [[ "$(id -u)" -ne 0 ]]; then
    die "This script must run as root."
  fi
}

VERSION=""
TAG=""
REPO_URL="https://github.com/jvxis/brln-os-light.git"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      VERSION="${2:-}"
      shift 2
      ;;
    --version=*)
      VERSION="${1#*=}"
      shift
      ;;
    --tag)
      TAG="${2:-}"
      shift 2
      ;;
    --tag=*)
      TAG="${1#*=}"
      shift
      ;;
    --repo-url)
      REPO_URL="${2:-}"
      shift 2
      ;;
    --repo-url=*)
      REPO_URL="${1#*=}"
      shift
      ;;
    *)
      die "Unknown argument: $1"
      ;;
  esac
done

require_root

VERSION="${VERSION#v}"
VERSION="$(echo "${VERSION}" | tr -s '[:space:]' ' ' | sed 's/^ *//;s/ *$//;s/ /-/g')"
if [[ -z "$VERSION" ]]; then
  die "Missing --version."
fi
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+([\-][0-9A-Za-z][0-9A-Za-z\.-]*)?$ ]]; then
  die "Invalid version format: ${VERSION}"
fi

if [[ -z "$TAG" ]]; then
  TAG="v${VERSION}"
fi
if [[ "$TAG" =~ [[:space:]] ]]; then
  die "Invalid tag format."
fi

LOCK_FILE="/var/lib/lightningos/app-upgrade.lock"
exec 9>"$LOCK_FILE"
if ! flock -n 9; then
  die "Another app upgrade is already running."
fi

for dep in git go npm systemctl install cp rm flock; do
  if ! command -v "$dep" >/dev/null 2>&1; then
    die "Required command missing: ${dep}"
  fi
done

mirror_root="/var/lib/lightningos/src/brln-os-light"
worktree_root="/var/lib/lightningos/worktrees"
worktree_dir=""

cleanup() {
  if [[ -n "$worktree_dir" && -d "$worktree_dir" ]]; then
    git -C "$mirror_root" worktree remove --force "$worktree_dir" >/dev/null 2>&1 || true
    rm -rf "$worktree_dir" >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT

print_step "Preparing repository mirror"
mkdir -p "$(dirname "$mirror_root")" "$worktree_root"
if [[ ! -d "$mirror_root/.git" ]]; then
  rm -rf "$mirror_root"
  git clone --no-checkout "$REPO_URL" "$mirror_root"
else
  git -C "$mirror_root" remote set-url origin "$REPO_URL" || true
fi
git -C "$mirror_root" fetch --tags --prune origin

if ! git -C "$mirror_root" rev-parse -q --verify "refs/tags/$TAG^{}" >/dev/null 2>&1; then
  alt_tag=""
  while IFS= read -r candidate; do
    if [[ "${candidate,,}" == "${TAG,,}" ]]; then
      alt_tag="$candidate"
      break
    fi
  done < <(git -C "$mirror_root" tag --list)
  if [[ -n "$alt_tag" ]]; then
    TAG="$alt_tag"
  fi
fi

if ! git -C "$mirror_root" rev-parse -q --verify "refs/tags/$TAG^{}" >/dev/null 2>&1; then
  die "Tag not found in repository: ${TAG}"
fi

print_step "Creating temporary worktree for ${TAG}"
safe_tag="$(echo "$TAG" | tr '/\\' '__')"
worktree_dir="${worktree_root}/app-upgrade-${safe_tag}-$(date +%Y%m%d%H%M%S)"
git -C "$mirror_root" worktree add --detach "$worktree_dir" "$TAG"

go_env="GOPATH=/opt/lightningos/go GOCACHE=/opt/lightningos/go-cache GOMODCACHE=/opt/lightningos/go/pkg/mod"
mkdir -p /opt/lightningos/go /opt/lightningos/go-cache /opt/lightningos/go/pkg/mod

print_step "Downloading Go modules"
(cd "$worktree_dir" && env $go_env GOFLAGS=-mod=mod go mod download)
print_ok "Go modules ready"

print_step "Building manager binary"
(cd "$worktree_dir" && env $go_env GOFLAGS=-mod=mod go build -o dist/lightningos-manager ./cmd/lightningos-manager)
install -m 0755 "$worktree_dir/dist/lightningos-manager" /opt/lightningos/manager/lightningos-manager
print_ok "Manager installed"

print_step "Building UI"
(cd "$worktree_dir/ui" && npm install && npm run build)
rm -rf /opt/lightningos/ui/*
cp -a "$worktree_dir/ui/dist/." /opt/lightningos/ui/
print_ok "UI installed"

print_step "Restarting lightningos-manager"
systemctl restart lightningos-manager
if systemctl is-active --quiet lightningos-manager; then
  print_ok "lightningos-manager is active"
else
  die "lightningos-manager failed to start"
fi

print_ok "App upgrade complete to ${VERSION} (${TAG})"
