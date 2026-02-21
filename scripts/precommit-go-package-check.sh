#!/usr/bin/env bash
set -euo pipefail

MODE="${1:-}"
if [[ -z "${MODE}" ]]; then
  echo "usage: $0 <vet|test|lint>" >&2
  exit 2
fi

if ! command -v git >/dev/null 2>&1; then
  echo "git is required" >&2
  exit 1
fi

if ! command -v go >/dev/null 2>&1; then
  echo "go is required" >&2
  exit 1
fi

ROOT_DIR="$(git rev-parse --show-toplevel)"
cd "${ROOT_DIR}"

mkdir -p "${ROOT_DIR}/.cache/go-build" "${ROOT_DIR}/.cache/go-mod"
export GOCACHE="${GOCACHE:-${ROOT_DIR}/.cache/go-build}"
export GOMODCACHE="${GOMODCACHE:-${ROOT_DIR}/.cache/go-mod}"
export GOLANGCI_LINT_CACHE="${GOLANGCI_LINT_CACHE:-${ROOT_DIR}/.cache/golangci-lint}"

STAGED_GO_FILES="$(git diff --cached --name-only --diff-filter=ACMR | grep -E '\.go$' || true)"
if [[ -z "${STAGED_GO_FILES}" ]]; then
  echo "No staged Go files, skipping ${MODE}"
  exit 0
fi

PKG_LINES=""
while IFS= read -r file; do
  if [[ -z "${file}" ]]; then
    continue
  fi
  if [[ ! -f "${file}" ]]; then
    continue
  fi

  dir="$(dirname "${file}")"
  pkg="./${dir}"
  if [[ "${dir}" == "." ]]; then
    pkg="./"
  fi

  if go list "${pkg}" >/dev/null 2>&1; then
    PKG_LINES+="${pkg}"$'\n'
  fi
done <<< "${STAGED_GO_FILES}"

PKGS="$(printf '%s' "${PKG_LINES}" | awk 'NF' | sort -u)"

if [[ -z "${PKGS}" ]]; then
  echo "No resolvable staged Go packages, skipping ${MODE}"
  exit 0
fi

echo "Running ${MODE} for packages:"
echo "${PKGS}"

case "${MODE}" in
  vet)
    while IFS= read -r pkg; do
      [[ -z "${pkg}" ]] && continue
      go vet "${pkg}"
    done <<< "${PKGS}"
    ;;
  test)
    while IFS= read -r pkg; do
      [[ -z "${pkg}" ]] && continue
      go test -short "${pkg}"
    done <<< "${PKGS}"
    ;;
  lint)
    mkdir -p "${GOLANGCI_LINT_CACHE}"
    LINT_BIN="${GOLANGCI_LINT:-$(go env GOPATH)/bin/golangci-lint}"
    if [[ -x "${LINT_BIN}" ]]; then
      :
    elif command -v golangci-lint >/dev/null 2>&1; then
      LINT_BIN="golangci-lint"
    else
      echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" >&2
      exit 1
    fi

    PKG_ARGS=()
    while IFS= read -r pkg; do
      [[ -z "${pkg}" ]] && continue
      PKG_ARGS+=("${pkg}")
    done <<< "${PKGS}"

    "${LINT_BIN}" run "${PKG_ARGS[@]}"
    ;;
  *)
    echo "unknown mode: ${MODE}" >&2
    exit 2
    ;;
esac
