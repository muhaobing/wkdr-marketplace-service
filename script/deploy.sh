#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BACKEND_PORT=10302
FRONTEND_PORT=10301
BIN_DIR="${PROJECT_DIR}/bin"
LOG_DIR="${PROJECT_DIR}/logs"
TARGET="${1:-all}"

echo "[deploy] project root: ${PROJECT_DIR}"

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

install_pkg() {
  local pkg="$1"
  if command_exists apt-get; then
    sudo apt-get update -y && sudo apt-get install -y "${pkg}"
  elif command_exists yum; then
    sudo yum install -y "${pkg}"
  elif command_exists dnf; then
    sudo dnf install -y "${pkg}"
  elif command_exists apk; then
    sudo apk add --no-cache "${pkg}"
  elif command_exists brew; then
    brew install "${pkg}"
  else
    echo "[deploy] no supported package manager found, please install ${pkg} manually"
    exit 1
  fi
}

ensure_dependencies() {
  echo "[deploy] checking dependencies..."
  command_exists lsof || install_pkg lsof
}

ensure_backend_dependencies() {
  command_exists go || install_pkg golang
}

ensure_frontend_dependencies() {
  command_exists node || install_pkg nodejs
  command_exists npm || install_pkg npm
}

kill_port() {
  local port="$1"
  local pids
  pids="$(lsof -ti :"${port}" || true)"
  if [[ -n "${pids}" ]]; then
    echo "[deploy] stopping processes on port ${port}: ${pids}"
    kill -9 ${pids}
  fi
}

prepare_backend() {
  echo "[deploy] preparing backend..."
  cd "${PROJECT_DIR}"
  export GOPROXY="https://goproxy.cn,direct"
  export GOFLAGS="-mod=mod"
  echo "[deploy] GOPROXY=${GOPROXY} GOFLAGS=${GOFLAGS}"
  go mod tidy
  go mod download
  go mod verify
  mkdir -p "${BIN_DIR}" "${LOG_DIR}"
  go build -o "${BIN_DIR}/wkdr-marketplace-service" ./cmd/main.go
}

prepare_frontend() {
  echo "[deploy] preparing frontend..."
  cd "${PROJECT_DIR}/web"
  if [[ ! -d node_modules ]]; then
    npm install
  else
    npm install --prefer-offline
  fi
}

start_backend() {
  echo "[deploy] starting backend on :${BACKEND_PORT} ..."
  cd "${PROJECT_DIR}"
  nohup "${BIN_DIR}/wkdr-marketplace-service" > "${LOG_DIR}/backend.log" 2>&1 &
  echo "[deploy] backend started, log: ${LOG_DIR}/backend.log"
}

start_frontend() {
  echo "[deploy] starting frontend on :${FRONTEND_PORT} ..."
  cd "${PROJECT_DIR}/web"
  nohup npm run dev -- --host 0.0.0.0 --port "${FRONTEND_PORT}" > "${LOG_DIR}/frontend.log" 2>&1 &
  echo "[deploy] frontend started, log: ${LOG_DIR}/frontend.log"
}

main() {
  ensure_dependencies

  case "${TARGET}" in
    be)
      ensure_backend_dependencies
      kill_port "${BACKEND_PORT}"
      prepare_backend
      start_backend
      ;;
    fe)
      ensure_frontend_dependencies
      kill_port "${FRONTEND_PORT}"
      mkdir -p "${LOG_DIR}"
      prepare_frontend
      start_frontend
      ;;
    all)
      ensure_backend_dependencies
      ensure_frontend_dependencies
      kill_port "${BACKEND_PORT}"
      kill_port "${FRONTEND_PORT}"
      prepare_backend
      prepare_frontend
      start_backend
      start_frontend
      ;;
    *)
      echo "[deploy] invalid arg: ${TARGET}"
      echo "[deploy] usage: bash script/deploy.sh [be|fe|all]"
      exit 1
      ;;
  esac

  echo "[deploy] done."
  if [[ "${TARGET}" == "all" || "${TARGET}" == "fe" ]]; then
    echo "[deploy] frontend: http://<your-host>:${FRONTEND_PORT}"
  fi
  if [[ "${TARGET}" == "all" || "${TARGET}" == "be" ]]; then
    echo "[deploy] backend : http://<your-host>:${BACKEND_PORT}"
  fi
}

main "$@"