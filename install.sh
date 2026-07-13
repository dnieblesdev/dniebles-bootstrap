#!/bin/bash
set -euo pipefail

# Direct binary installer for dbootstrap on Linux/WSL amd64 and arm64.
#
# Usage: ./install.sh [--version vX.Y.Z] [--allow-prerelease] [--force] [--uninstall]
#
# The script is intentionally auditable: it resolves one GitHub Release,
# downloads the archive and matching SHA-256 file, verifies the checksum
# before extraction, validates the payload, then atomically installs the
# binary and catalog. No package manager or sudo is used.

INSTALLER_API_BASE="${INSTALLER_API_BASE:-https://api.github.com/repos/dnieblesdev/dniebles-bootstrap}"
INSTALLER_DOWNLOAD_BASE="${INSTALLER_DOWNLOAD_BASE:-https://github.com/dnieblesdev/dniebles-bootstrap/releases/download}"
INSTALLER_HTTP_CMD="${INSTALLER_HTTP_CMD:-curl -fsSL}"
INSTALLER_SHA256_CMD="${INSTALLER_SHA256_CMD:-sha256sum --check --status --strict}"
INSTALLER_TAR_CMD="${INSTALLER_TAR_CMD:-tar}"
INSTALLER_UNAME_S="${INSTALLER_UNAME_S:-$(uname -s)}"
INSTALLER_UNAME_M="${INSTALLER_UNAME_M:-$(uname -m)}"

REPO_OWNER="dnieblesdev"
REPO_NAME="dniebles-bootstrap"

BINARY_NAME="dbootstrap"
CATALOG_NAME="bootstrap.toml"
STATE_NAME="install-state.toml"

main() {
  local requested_version=""
  local allow_prerelease=false
  local force=false
  local uninstall=false

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --version)
        requested_version="${2:-}"
        if [[ -z "$requested_version" ]]; then
          die "error: --version requires a value"
        fi
        shift 2
        ;;
      --allow-prerelease)
        allow_prerelease=true
        shift
        ;;
      --force)
        force=true
        shift
        ;;
      --uninstall)
        uninstall=true
        shift
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        die "error: unknown option $1"
        ;;
    esac
  done

  local bin_dir catalog_dir data_dir
  bin_dir="${XDG_BIN_HOME:-${HOME}/.local/bin}"
  data_dir="${XDG_DATA_HOME:-${HOME}/.local/share}/dbootstrap"
  catalog_dir="${data_dir}/catalog"

  local binary_path catalog_path state_path
  binary_path="${bin_dir}/${BINARY_NAME}"
  catalog_path="${catalog_dir}/${CATALOG_NAME}"
  state_path="${data_dir}/${STATE_NAME}"

  if [[ "$uninstall" == true ]]; then
    do_uninstall "$binary_path" "$catalog_path" "$state_path"
    return 0
  fi

  local os arch
  detect_platform "$INSTALLER_UNAME_S" "$INSTALLER_UNAME_M"
  os="$DETECTED_OS"
  arch="$DETECTED_ARCH"

  ensure_directory "$bin_dir"
  ensure_directory "$catalog_dir"

  local release_json
  release_json="$(resolve_release "$requested_version" "$allow_prerelease")"

  local tag prerelease safe_version
  tag="$(python3 -c "import sys, json; print(json.load(sys.stdin)['tag_name'])" <<< "$release_json")"
  prerelease="$(python3 -c "import sys, json; print(str(json.load(sys.stdin)['prerelease']).lower())" <<< "$release_json")"
  safe_version="$(safe_version_from_tag "$tag")"

  if [[ "$prerelease" == "true" && "$allow_prerelease" != true ]]; then
    die "error: $tag is a prerelease; use --allow-prerelease to install it"
  fi

  local archive_name archive_url checksum_url
  archive_name="dbootstrap_${safe_version}_${os}_${arch}.tar.gz"
  archive_url="${INSTALLER_DOWNLOAD_BASE}/download/${tag}/${archive_name}"
  checksum_url="${archive_url}.sha256"

  work_dir="$(mktemp -d)"
  trap 'rm -rf "$work_dir"' EXIT

  local archive_file checksum_file
  archive_file="${work_dir}/${archive_name}"
  checksum_file="${work_dir}/${archive_name}.sha256"

  download "$archive_url" "$archive_file"
  download "$checksum_url" "$checksum_file"

  verify_checksum "$checksum_file" "$work_dir"

  local extract_dir
  extract_dir="${work_dir}/extract"
  mkdir -p "$extract_dir"
  extract_archive "$archive_file" "$extract_dir"
  validate_payload "$extract_dir"

  local staged_binary staged_catalog
  staged_binary="${extract_dir}/${BINARY_NAME}"
  staged_catalog="${extract_dir}/catalog/${CATALOG_NAME}"

  local staged_binary_digest staged_catalog_digest
  staged_binary_digest="sha256:$(file_digest "$staged_binary")"
  staged_catalog_digest="sha256:$(file_digest "$staged_catalog")"

  recover_or_cleanup_transaction "$binary_path" "$catalog_path" "$state_path" "$(transaction_dir_for "$data_dir")"

  assert_installable "$binary_path" "$catalog_path" "$state_path" "$force"

  local tx_dir
  tx_dir="$(transaction_dir_for "$data_dir")"
  begin_transaction "$tx_dir" "$state_path" "$binary_path" "$catalog_path" "$tag" "$staged_binary_digest" "$staged_catalog_digest"

  if ! commit_transaction "$tx_dir" "$staged_binary" "$binary_path" "$staged_catalog" "$catalog_path" "$state_path" "$tag" "$staged_binary_digest" "$staged_catalog_digest"; then
    rollback_transaction "$tx_dir" "$binary_path" "$catalog_path" "$state_path"
    die "error: installation failed; previous state restored"
  fi

  if ! directory_on_path "$bin_dir"; then
    echo ""
    echo "Add ${bin_dir} to your PATH, for example:"
    echo "  export PATH=\"${bin_dir}:\$PATH\""
  fi

  echo "Installed dbootstrap ${tag}:"
  echo "  binary:  ${binary_path}"
  echo "  catalog: ${catalog_path}"
}

usage() {
  cat <<'USAGE'
Usage: ./install.sh [options]

Options:
  --version vX.Y.Z       Install a specific release tag
  --allow-prerelease     Allow installation of prerelease tags
  --force                Allow reinstall, upgrade, or downgrade
  --uninstall            Remove the managed binary, catalog, and state
  -h, --help             Show this help message
USAGE
}

die() {
  echo "$1" >&2
  exit 1
}

DETECTED_OS=""
DETECTED_ARCH=""

detect_platform() {
  local raw_os="$1"
  local raw_arch="$2"

  local os
  os="$(printf '%s' "$raw_os" | tr '[:upper:]' '[:lower:]')"

  case "$os" in
    linux)
      DETECTED_OS="linux"
      ;;
    *)
      die "error: unsupported OS '${raw_os}'; direct binary installation is available only for Linux and WSL"
      ;;
  esac

  case "$raw_arch" in
    x86_64|amd64)
      DETECTED_ARCH="amd64"
      ;;
    aarch64|arm64)
      DETECTED_ARCH="arm64"
      ;;
    *)
      die "error: unsupported architecture '${raw_arch}'; direct binary installation supports only amd64 and arm64"
      ;;
  esac
}

ensure_directory() {
  if [[ ! -d "$1" ]]; then
    mkdir -p "$1"
  fi
}

resolve_release() {
  local version="$1"
  local allow_prerelease="$2"
  local url

  if [[ -n "$version" ]]; then
    url="${INSTALLER_API_BASE}/releases/tags/${version}"
  else
    url="${INSTALLER_API_BASE}/releases/latest"
  fi

  local response
  response="$($INSTALLER_HTTP_CMD "$url" 2>&1)" || die "error: failed to resolve release from ${url}: ${response}"

  if ! python3 -c "import sys, json; json.load(sys.stdin)['tag_name']" <<< "$response" >/dev/null 2>&1; then
    die "error: release response did not contain a valid tag_name"
  fi

  if [[ -n "$version" && "$allow_prerelease" != true ]]; then
    local actual_tag
    actual_tag="$(python3 -c "import sys, json; print(json.load(sys.stdin)['tag_name'])" <<< "$response")"
    if [[ "$actual_tag" != "$version" ]]; then
      die "error: requested ${version} but API returned ${actual_tag}"
    fi
  fi

  printf '%s' "$response"
}

safe_version_from_tag() {
  local tag="$1"
  # Mirror NormalizeGitVersion: keep [A-Za-z0-9._-], collapse invalid runs to '-', trim separators.
  local safe
  safe="$(printf '%s' "$tag" | sed 's/[^A-Za-z0-9._-]/-/g; s/^[-_.]*//; s/[-_.]*$//')"
  if [[ -z "$safe" ]]; then
    safe="dev"
  fi
  printf '%s' "$safe"
}

download() {
  local url="$1"
  local dest="$2"
  $INSTALLER_HTTP_CMD "$url" > "$dest" 2>/dev/null || die "error: failed to download ${url}"
}

verify_checksum() {
  local checksum_file="$1"
  local work_dir="$2"
  (cd "$work_dir" && $INSTALLER_SHA256_CMD "$checksum_file") || die "error: checksum verification failed; archive may be corrupted or tampered"
}

extract_archive() {
  local archive="$1"
  local dest="$2"
  $INSTALLER_TAR_CMD -xzf "$archive" -C "$dest"
}

validate_payload() {
  local dir="$1"
  if [[ ! -f "${dir}/${BINARY_NAME}" ]]; then
    die "error: archive is missing the dbootstrap binary"
  fi
  if [[ ! -f "${dir}/catalog/${CATALOG_NAME}" ]]; then
    die "error: archive is missing catalog/bootstrap.toml"
  fi
}

file_digest() {
  local file="$1"
  sha256sum "$file" | awk '{print $1}'
}

transaction_dir_for() {
  local data_dir="$1"
  printf '%s' "${data_dir}/.install-tx"
}

files_match_state() {
  local binary_path="$1"
  local catalog_path="$2"
  local state_path="$3"

  if [[ ! -f "$binary_path" || ! -f "$catalog_path" ]]; then
    return 1
  fi

  local current_binary current_catalog expected_binary expected_catalog
  current_binary="sha256:$(file_digest "$binary_path")"
  current_catalog="sha256:$(file_digest "$catalog_path")"
  expected_binary="$(state_digest_for_path "$state_path" "$binary_path")"
  expected_catalog="$(state_digest_for_path "$state_path" "$catalog_path")"

  if [[ -z "$expected_binary" || -z "$expected_catalog" ]]; then
    return 1
  fi

  [[ "$current_binary" == "$expected_binary" && "$current_catalog" == "$expected_catalog" ]]
}

begin_transaction() {
  local tx_dir="$1"
  local state_path="$2"
  local binary_path="$3"
  local catalog_path="$4"
  local tag="$5"
  local binary_digest="$6"
  local catalog_digest="$7"

  rm -rf "$tx_dir"
  mkdir -p "${tx_dir}/backup"

  if [[ -f "$binary_path" ]]; then
    cp -p "$binary_path" "${tx_dir}/backup/bin"
  fi
  if [[ -f "$catalog_path" ]]; then
    cp -p "$catalog_path" "${tx_dir}/backup/catalog"
  fi
  if [[ -f "$state_path" ]]; then
    cp -p "$state_path" "${tx_dir}/backup/state"
  fi

  write_state "${tx_dir}/intended-state.toml" "$tag" "$binary_path" "$binary_digest" "$catalog_path" "$catalog_digest"
}

commit_transaction() {
  local tx_dir="$1"
  local staged_binary="$2"
  local binary_path="$3"
  local staged_catalog="$4"
  local catalog_path="$5"
  local state_path="$6"
  local tag="$7"
  local binary_digest="$8"
  local catalog_digest="$9"

  atomic_replace "$staged_binary" "$binary_path" || return 1
  atomic_replace "$staged_catalog" "$catalog_path" || return 1

  write_state "${state_path}.tmp" "$tag" "$binary_path" "$binary_digest" "$catalog_path" "$catalog_digest" || return 1
  mv "${state_path}.tmp" "$state_path" || return 1

  rm -rf "$tx_dir"
}

rollback_transaction() {
  local tx_dir="$1"
  local binary_path="$2"
  local catalog_path="$3"
  local state_path="$4"

  if [[ ! -d "$tx_dir" ]]; then
    return 0
  fi

  if [[ -f "${tx_dir}/backup/bin" ]]; then
    cp -p "${tx_dir}/backup/bin" "$binary_path"
  else
    rm -f "$binary_path"
  fi

  if [[ -f "${tx_dir}/backup/catalog" ]]; then
    cp -p "${tx_dir}/backup/catalog" "$catalog_path"
  else
    rm -f "$catalog_path"
  fi

  if [[ -f "${tx_dir}/backup/state" ]]; then
    cp -p "${tx_dir}/backup/state" "$state_path"
  else
    rm -f "$state_path"
  fi

  rm -rf "$tx_dir"
}

recover_or_cleanup_transaction() {
  local binary_path="$1"
  local catalog_path="$2"
  local state_path="$3"
  local tx_dir="$4"

  if [[ ! -d "$tx_dir" ]]; then
    return 0
  fi

  local committed=false
  local intended_state="${tx_dir}/intended-state.toml"

  if [[ -f "$intended_state" && -f "$state_path" ]]; then
    if cmp -s "$intended_state" "$state_path"; then
      committed=true
    fi
  fi

  if [[ "$committed" != true && -f "$state_path" ]]; then
    if files_match_state "$binary_path" "$catalog_path" "$state_path"; then
      committed=true
    fi
  fi

  if [[ "$committed" == true ]]; then
    rm -rf "$tx_dir"
    return 0
  fi

  rollback_transaction "$tx_dir" "$binary_path" "$catalog_path" "$state_path"
}

validate_state_ownership() {
  local binary_path="$1"
  local catalog_path="$2"
  local state_path="$3"

  if [[ ! -f "$state_path" ]]; then
    echo "error: no install state found at ${state_path}" >&2
    return 1
  fi

  local release
  release="$(awk '/^release *= */ { gsub(/^release *= *"|"$/, ""); print; exit }' "$state_path")"
  if [[ -z "$release" ]]; then
    echo "error: install state is missing release" >&2
    return 1
  fi

  local entries
  if ! entries="$(awk '
    BEGIN { count=0; valid=1 }
    /^\[\[managed\]\]/ {
      count++
      in_block=1
      path=""
      digest=""
      next
    }
    in_block && /^path *= */ {
      gsub(/^path *= *"|"$/, "", $0)
      path=$0
      next
    }
    in_block && /^digest *= */ {
      gsub(/^digest *= *"|"$/, "", $0)
      digest=$0
      next
    }
    /^$/ {
      if (in_block) {
        if (path == "" || digest == "") valid=0
        else print path "\t" digest
      }
      in_block=0
    }
    END {
      if (in_block) {
        if (path == "" || digest == "") valid=0
        else print path "\t" digest
      }
      if (count != 2 || !valid) exit 1
    }
  ' "$state_path")"; then
    echo "error: install state managed section is malformed" >&2
    return 1
  fi

  local managed_paths=()
  local managed_digests=()
  local line path digest
  while IFS=$'\t' read -r path digest; do
    managed_paths+=("$path")
    managed_digests+=("$digest")
  done <<< "$entries"

  if [[ ${#managed_paths[@]} -ne 2 ]]; then
    echo "error: install state must contain exactly two managed paths, found ${#managed_paths[@]}" >&2
    return 1
  fi

  local sorted_expected sorted_found
  sorted_expected="$(printf '%s\n' "$binary_path" "$catalog_path" | sort)"
  sorted_found="$(printf '%s\n' "${managed_paths[@]}" | sort)"
  if [[ "$sorted_expected" != "$sorted_found" ]]; then
    echo "error: install state managed paths do not match expected paths" >&2
    return 1
  fi

  local i actual_digest
  for i in 0 1; do
    path="${managed_paths[$i]}"
    digest="${managed_digests[$i]}"
    if [[ ! -f "$path" ]]; then
      echo "error: managed file is missing: ${path}" >&2
      return 1
    fi
    actual_digest="sha256:$(file_digest "$path")"
    if [[ "$actual_digest" != "$digest" ]]; then
      echo "error: managed file digest mismatch at ${path}; refusing force install" >&2
      return 1
    fi
  done
}

assert_installable() {
  local binary_path="$1"
  local catalog_path="$2"
  local state_path="$3"
  local force="$4"

  # Unmanaged files abort regardless of --force; only the manifest owns managed paths.
  if [[ -e "$binary_path" && ! -f "$state_path" ]]; then
    die "error: unmanaged file exists at ${binary_path}; remove it or install elsewhere"
  fi
  if [[ -e "$catalog_path" && ! -f "$state_path" ]]; then
    die "error: unmanaged file exists at ${catalog_path}; remove it or install elsewhere"
  fi

  # No state means no managed install; proceed (directories already ensured).
  if [[ ! -f "$state_path" ]]; then
    return 0
  fi

  # A matching managed installation requires explicit force for reinstall/upgrade/downgrade.
  if [[ "$force" != true ]]; then
    die "error: dbootstrap is already installed; use --force to reinstall, upgrade, or downgrade"
  fi

  # With --force, the manifest must be fully trusted: parseable, exact paths, matching digests.
  validate_state_ownership "$binary_path" "$catalog_path" "$state_path" || die "error: install state is not trusted; aborting force install"
}

atomic_replace() {
  local source="$1"
  local target="$2"
  local tmp="${target}.install-tmp"

  if cp "$source" "$tmp" && mv "$tmp" "$target"; then
    return 0
  fi
  rm -f "$tmp"
  return 1
}

write_state() {
  local state_path="$1"
  local tag="$2"
  local binary_path="$3"
  local binary_digest="$4"
  local catalog_path="$5"
  local catalog_digest="$6"

  cat > "$state_path" <<EOF
release = "${tag}"

[[managed]]
path = "${binary_path}"
digest = "${binary_digest}"

[[managed]]
path = "${catalog_path}"
digest = "${catalog_digest}"
EOF
}

do_uninstall() {
  local binary_path="$1"
  local catalog_path="$2"
  local state_path="$3"

  if [[ ! -f "$state_path" ]]; then
    die "error: no install state found at ${state_path}; refusing to uninstall"
  fi

  local missing=()
  for file in "$binary_path" "$catalog_path"; do
    if [[ ! -f "$file" ]]; then
      missing+=("$file")
    fi
  done

  if [[ ${#missing[@]} -gt 0 ]]; then
    die "error: managed files are missing: ${missing[*]}; refusing to uninstall"
  fi

  local current_binary current_catalog
  current_binary="sha256:$(file_digest "$binary_path")"
  current_catalog="sha256:$(file_digest "$catalog_path")"

  local expected_binary expected_catalog
  expected_binary="$(state_digest_for_path "$state_path" "$binary_path")"
  expected_catalog="$(state_digest_for_path "$state_path" "$catalog_path")"

  if [[ -z "$expected_binary" || -z "$expected_catalog" ]]; then
    die "error: install state is malformed; refusing to uninstall"
  fi

  local modified=()
  if [[ "$current_binary" != "$expected_binary" ]]; then
    modified+=("$binary_path")
  fi
  if [[ "$current_catalog" != "$expected_catalog" ]]; then
    modified+=("$catalog_path")
  fi

  if [[ ${#modified[@]} -gt 0 ]]; then
    die "error: managed files have been modified: ${modified[*]}; aborting uninstall to preserve your changes"
  fi

  rm -f "$binary_path" "$catalog_path" "$state_path"
  rmdir "$(dirname "$catalog_path")" 2>/dev/null || true
  rmdir "$(dirname "$state_path")" 2>/dev/null || true
  echo "Uninstalled dbootstrap:"
  echo "  removed ${binary_path}"
  echo "  removed ${catalog_path}"
  echo "  removed ${state_path}"
}

state_digest_for_path() {
  local state_path="$1"
  local target_path="$2"
  awk -v p="$target_path" '
    /^\[\[managed\]\]/ { in_block=1; matched=0; next }
    in_block && /^path *= */ {
      gsub(/^path *= *"|"$/, "", $0)
      if ($0 == p) matched=1
    }
    in_block && /^digest *= */ && matched {
      gsub(/^digest *= *"|"$/, "", $0)
      print
      exit
    }
    /^$/ { in_block=0; matched=0 }
  ' "$state_path"
}

directory_on_path() {
  local dir="$1"
  case ":${PATH}:" in
    *:"$dir":*)
      return 0
      ;;
  esac
  return 1
}

main "$@"
