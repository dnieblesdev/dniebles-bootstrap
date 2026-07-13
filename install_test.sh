#!/bin/bash
set -euo pipefail

INSTALLER="${INSTALLER:-./install.sh}"
FAILED=0

fail() {
  echo "FAIL: $1"
  FAILED=1
}

assert_eq() {
  if [[ "$1" != "$2" ]]; then
    fail "expected $3 to be '$2', got '$1'"
  fi
}

assert_contains() {
  if [[ ! "$1" == *"$2"* ]]; then
    fail "expected $3 to contain '$2'; got '$1'"
  fi
}

assert_file_exists() {
  if [[ ! -f "$1" ]]; then
    fail "expected file '$1' to exist"
  fi
}

assert_file_missing() {
  if [[ -e "$1" ]]; then
    fail "expected file '$1' to be missing"
  fi
}

file_digest_local() {
  sha256sum "$1" | awk '{print $1}'
}

make_home() {
  local home="$1"
  mkdir -p "$home/.local/bin"
  mkdir -p "$home/.local/share/dbootstrap/catalog"
}

build_archive() {
  local staging="$1"
  local archive="$2"
  mkdir -p "$staging/catalog"
  echo '#!/bin/sh' > "$staging/dbootstrap"
  echo 'echo "dbootstrap mock"' >> "$staging/dbootstrap"
  chmod +x "$staging/dbootstrap"
  cat > "$staging/catalog/bootstrap.toml" <<'TOML'
schema = "dniebles.catalog"
version = 1

[[tools]]
id = "git"
description = "Version control"

[[profiles]]
id = "dev"
resources = ["tool:git"]
TOML
  tar -czf "$archive" -C "$staging" .
}

setup_fixtures() {
  local fixtures="$1"
  local version="$2"
  local prerelease="${3:-false}"
  local safe_version="${version}"
  local archive_name="dbootstrap_${safe_version}_linux_amd64.tar.gz"

  mkdir -p "$fixtures/api/releases/tags"
  mkdir -p "$fixtures/download/$version"

  cat > "$fixtures/api/releases/tags/$version" <<JSON
{
  "tag_name": "$version",
  "prerelease": $prerelease,
  "assets": [
    { "name": "$archive_name", "browser_download_url": "$fixtures/download/$version/$archive_name" },
    { "name": "$archive_name.sha256", "browser_download_url": "$fixtures/download/$version/$archive_name.sha256" }
  ]
}
JSON

  if [[ "$prerelease" != "true" ]]; then
    cp "$fixtures/api/releases/tags/$version" "$fixtures/api/releases/latest"
  fi

  local staging="$fixtures/staging-$version"
  build_archive "$staging" "$fixtures/download/$version/$archive_name"
  (cd "$fixtures/download/$version" && sha256sum "$archive_name" > "$archive_name.sha256")
}

capture() {
  local output
  output="$($@ 2>&1)" && echo "0:$output" || echo "$?:$output"
}

# RED: install.sh must exist and be valid Bash before any behavior tests.
test_installer_exists() {
  if [[ ! -f "$INSTALLER" ]]; then
    fail "installer '$INSTALLER' not found"
    return
  fi
  if ! bash -n "$INSTALLER"; then
    fail "installer fails bash syntax check"
  fi
}

# Unsupported platform is rejected without mutation.
test_unsupported_platform() {
  local home
  home="$(mktemp -d)"
  make_home "$home"
  local out
  out="$(INSTALLER_UNAME_S=Darwin INSTALLER_UNAME_M=x86_64 HOME="$home" "$INSTALLER" 2>&1)" && true
  local code=$?
  if [[ $code -eq 0 ]]; then
    fail "unsupported platform should fail"
  fi
  assert_contains "$out" "unsupported" "unsupported platform output"
  assert_file_missing "$home/.local/bin/dbootstrap"
}

# A supported install downloads, verifies, and places both managed files.
test_supported_install() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" 2>&1)" && code=0 || code=$?
  if [[ $code -ne 0 ]]; then
    fail "supported install should succeed; output: $out"
    return
  fi
  assert_file_exists "$home/.local/bin/dbootstrap"
  assert_file_exists "$home/.local/share/dbootstrap/catalog/bootstrap.toml"
  assert_file_exists "$home/.local/share/dbootstrap/install-state.toml"
  assert_contains "$out" "Installed" "install success output"
}

# Checksum mismatch aborts before extraction and leaves nothing new.
test_checksum_mismatch() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"
  echo "0000000000000000000000000000000000000000000000000000000000000000  dbootstrap_v1.2.3_linux_amd64.tar.gz" > "$fixtures/download/v1.2.3/dbootstrap_v1.2.3_linux_amd64.tar.gz.sha256"

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" 2>&1)" && code=0 || code=$?
  if [[ $code -eq 0 ]]; then
    fail "checksum mismatch should fail"
  fi
  assert_contains "$out" "checksum" "checksum mismatch output"
  assert_file_missing "$home/.local/bin/dbootstrap"
  assert_file_missing "$home/.local/share/dbootstrap/catalog/bootstrap.toml"
}

# An unmanaged file at a managed path aborts without overwrite.
test_unmanaged_file_refused() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"
  echo "unmanaged" > "$home/.local/bin/dbootstrap"

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" 2>&1)" && code=0 || code=$?
  if [[ $code -eq 0 ]]; then
    fail "unmanaged file should refuse install"
  fi
  assert_contains "$out" "unmanaged" "unmanaged file output"
}

# Reinstall of a matching managed install requires --force.
test_force_required_for_managed_reinstall() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"

  HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" >/dev/null 2>&1

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" 2>&1)" && code=0 || code=$?
  if [[ $code -eq 0 ]]; then
    fail "managed reinstall without force should fail"
  fi
  assert_contains "$out" "force" "managed reinstall output"

  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" --force 2>&1)" && code=0 || code=$?
  if [[ $code -ne 0 ]]; then
    fail "managed reinstall with --force should succeed; output: $out"
  fi
}

# Exact version selection is honored.
test_exact_version() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v2.0.0"

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" --version v2.0.0 2>&1)" && code=0 || code=$?
  if [[ $code -ne 0 ]]; then
    fail "exact version install should succeed; output: $out"
  fi
  assert_contains "$(cat "$home/.local/share/dbootstrap/install-state.toml")" "v2.0.0" "state records exact version"
}

# Prerelease requires --allow-prerelease.
test_prerelease_requires_flag() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.3.0-rc.1" true

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" --version v1.3.0-rc.1 2>&1)" && code=0 || code=$?
  if [[ $code -eq 0 ]]; then
    fail "prerelease without flag should fail"
  fi
  assert_contains "$out" "prerelease" "prerelease rejection output"

  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" --version v1.3.0-rc.1 --allow-prerelease 2>&1)" && code=0 || code=$?
  if [[ $code -ne 0 ]]; then
    fail "prerelease with flag should succeed; output: $out"
  fi
}

# Uninstall removes only unmodified managed files.
test_safe_uninstall() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"

  HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" >/dev/null 2>&1

  local out code
  out="$(HOME="$home" "$INSTALLER" --uninstall 2>&1)" && code=0 || code=$?
  if [[ $code -ne 0 ]]; then
    fail "uninstall should succeed; output: $out"
  fi
  assert_file_missing "$home/.local/bin/dbootstrap"
  assert_file_missing "$home/.local/share/dbootstrap/catalog/bootstrap.toml"
  assert_file_missing "$home/.local/share/dbootstrap/install-state.toml"
}

# Uninstall preserves modified managed files.
test_uninstall_preserves_modified() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"

  HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" >/dev/null 2>&1
  echo "modified" > "$home/.local/bin/dbootstrap"

  local out code
  out="$(HOME="$home" "$INSTALLER" --uninstall 2>&1)" && code=0 || code=$?
  if [[ $code -eq 0 ]]; then
    fail "uninstall with modified file should fail"
  fi
  assert_contains "$out" "modified" "modified file output"
  assert_file_exists "$home/.local/bin/dbootstrap"
}

# --force aborts when state manifest is malformed.
test_force_aborts_malformed_state() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"

  HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" >/dev/null 2>&1

  echo "not-valid-toml" > "$home/.local/share/dbootstrap/install-state.toml"

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" --force 2>&1)" && code=0 || code=$?
  if [[ $code -eq 0 ]]; then
    fail "--force with malformed state should fail"
  fi
  assert_contains "$out" "state" "malformed state output"
}

# --force aborts when managed paths in state do not match expected paths.
test_force_aborts_wrong_paths() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"

  HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" >/dev/null 2>&1

  cat > "$home/.local/share/dbootstrap/install-state.toml" <<EOF
release = "v1.2.3"

[[managed]]
path = "$home/.local/bin/dbootstrap"
digest = "sha256:$(file_digest_local "$home/.local/bin/dbootstrap")"

[[managed]]
path = "/some/other/path"
digest = "sha256:$(file_digest_local "$home/.local/share/dbootstrap/catalog/bootstrap.toml")"
EOF

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" --force 2>&1)" && code=0 || code=$?
  if [[ $code -eq 0 ]]; then
    fail "--force with wrong paths should fail"
  fi
  assert_contains "$out" "path" "wrong paths output"
}

# --force aborts when managed binary digest does not match manifest.
test_force_aborts_tampered_binary() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"

  HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" >/dev/null 2>&1
  echo "tampered" >> "$home/.local/bin/dbootstrap"

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" --force 2>&1)" && code=0 || code=$?
  if [[ $code -eq 0 ]]; then
    fail "--force with tampered binary should fail"
  fi
  assert_contains "$out" "digest" "tampered binary output"
}

# --force aborts when managed catalog digest does not match manifest.
test_force_aborts_tampered_catalog() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"

  HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" >/dev/null 2>&1
  echo "tampered" >> "$home/.local/share/dbootstrap/catalog/bootstrap.toml"

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" --force 2>&1)" && code=0 || code=$?
  if [[ $code -eq 0 ]]; then
    fail "--force with tampered catalog should fail"
  fi
  assert_contains "$out" "digest" "tampered catalog output"
}

# --force must never overwrite unmanaged files at managed paths.
test_force_does_not_overwrite_unmanaged() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"
  echo "unmanaged-binary" > "$home/.local/bin/dbootstrap"
  echo "unmanaged-catalog" > "$home/.local/share/dbootstrap/catalog/bootstrap.toml"

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" --force 2>&1)" && code=0 || code=$?
  if [[ $code -eq 0 ]]; then
    fail "--force should not overwrite unmanaged files"
  fi
  assert_contains "$out" "unmanaged" "force unmanaged output"
  assert_eq "$(cat "$home/.local/bin/dbootstrap")" "unmanaged-binary" "binary preserved under force"
  assert_eq "$(cat "$home/.local/share/dbootstrap/catalog/bootstrap.toml")" "unmanaged-catalog" "catalog preserved under force"
  assert_file_missing "$home/.local/share/dbootstrap/install-state.toml"
}

# A failure during cross-file replacement rolls back to the previous managed state.
test_transaction_rollback_on_failure() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"

  HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" >/dev/null 2>&1

  # Prepare a different version so we can detect partial mutation vs rollback.
  setup_fixtures "$fixtures" "v2.0.0"

  # Make the catalog directory read-only so catalog replacement fails after binary replacement.
  chmod 555 "$home/.local/share/dbootstrap/catalog"

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" --force 2>&1)" && code=0 || code=$?

  # Restore writability so the temp home can be cleaned up.
  chmod 755 "$home/.local/share/dbootstrap/catalog"

  if [[ $code -eq 0 ]]; then
    fail "transaction failure should abort install"
  fi
  assert_contains "$(cat "$home/.local/share/dbootstrap/install-state.toml")" "v1.2.3" "state rolled back to original"
  assert_contains "$(cat "$home/.local/bin/dbootstrap")" "dbootstrap mock" "binary rolled back to original"
}

# A retained transaction from an interrupted install is recovered before new work proceeds.
test_transaction_recovery_on_next_run() {
  local fixtures home
  fixtures="$(mktemp -d)"
  home="$(mktemp -d)"
  make_home "$home"
  setup_fixtures "$fixtures" "v1.2.3"

  HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" >/dev/null 2>&1

  setup_fixtures "$fixtures" "v2.0.0"

  # Simulate an interrupted upgrade: files replaced, state old, tx dir with backups.
  local data_dir tx_dir staging archive
  data_dir="$home/.local/share/dbootstrap"
  tx_dir="$data_dir/.install-tx"
  staging="$(mktemp -d)"
  archive="$fixtures/download/v2.0.0/dbootstrap_v2.0.0_linux_amd64.tar.gz"
  tar -xzf "$archive" -C "$staging"

  mkdir -p "$tx_dir/backup"
  cp -p "$home/.local/bin/dbootstrap" "$tx_dir/backup/bin"
  cp -p "$home/.local/share/dbootstrap/catalog/bootstrap.toml" "$tx_dir/backup/catalog"
  cp -p "$home/.local/share/dbootstrap/install-state.toml" "$tx_dir/backup/state"

  local bin_digest cat_digest
  bin_digest="sha256:$(file_digest_local "$staging/dbootstrap")"
  cat_digest="sha256:$(file_digest_local "$staging/catalog/bootstrap.toml")"

  cat > "$tx_dir/intended-state.toml" <<EOF
release = "v2.0.0"

[[managed]]
path = "$home/.local/bin/dbootstrap"
digest = "$bin_digest"

[[managed]]
path = "$home/.local/share/dbootstrap/catalog/bootstrap.toml"
digest = "$cat_digest"
EOF

  cp "$staging/dbootstrap" "$home/.local/bin/dbootstrap"
  cp "$staging/catalog/bootstrap.toml" "$home/.local/share/dbootstrap/catalog/bootstrap.toml"

  local out code
  out="$(HOME="$home" INSTALLER_API_BASE="file://$fixtures/api" INSTALLER_DOWNLOAD_BASE="file://$fixtures" "$INSTALLER" --force 2>&1)" && code=0 || code=$?
  if [[ $code -ne 0 ]]; then
    fail "recovery install should succeed; output: $out"
  fi
  assert_contains "$(cat "$home/.local/share/dbootstrap/install-state.toml")" "v2.0.0" "state reflects new version after recovery"
}

main() {
  test_installer_exists
  test_unsupported_platform
  test_supported_install
  test_checksum_mismatch
  test_unmanaged_file_refused
  test_force_required_for_managed_reinstall
  test_force_aborts_malformed_state
  test_force_aborts_wrong_paths
  test_force_aborts_tampered_binary
  test_force_aborts_tampered_catalog
  test_exact_version
  test_prerelease_requires_flag
  test_safe_uninstall
  test_uninstall_preserves_modified
  test_force_does_not_overwrite_unmanaged
  test_transaction_rollback_on_failure
  test_transaction_recovery_on_next_run

  if [[ $FAILED -ne 0 ]]; then
    echo "Some tests failed."
    exit 1
  fi
  echo "All install tests passed."
}

main "$@"
