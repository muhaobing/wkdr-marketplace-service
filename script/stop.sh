#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

BACKEND_PORT=10302
FRONTEND_PORT=10301
TARGET="${1:-all}"

echo "[stop] project root: ${PROJECT_DIR}"

kill_port() {
  local port="$1"
  local pids=""

  if command -v lsof >/dev/null 2>&1; then
    pids="$(lsof -ti :"${port}" || true)"
  elif command -v ss >/dev/null 2>&1; then
    pids="$(ss -ltnp "( sport = :${port} )" 2>/dev/null | awk -F 'pid=' 'NR>1 {print $2}' | awk -F ',' '{print $1}' | sort -u || true)"
  fi

  if [[ -z "${pids}" ]]; then
    echo "[stop] no process found on port ${port}"
    return
  fi

  echo "[stop] stopping processes on port ${port}: ${pids}"
  kill -9 ${pids}
}

main() {
  case "${TARGET}" in
    be)
      kill_port "${BACKEND_PORT}"
      ;;
    fe)
      kill_port "${FRONTEND_PORT}"
      ;;
    all)
      kill_port "${BACKEND_PORT}"
      kill_port "${FRONTEND_PORT}"
      ;;
    *)
      echo "[stop] invalid arg: ${TARGET}"
      echo "[stop] usage: bash script/stop.sh [be|fe|all]"
      exit 1
      ;;
  esac

  echo "[stop] done. target=${TARGET}"
}

main "$@"
