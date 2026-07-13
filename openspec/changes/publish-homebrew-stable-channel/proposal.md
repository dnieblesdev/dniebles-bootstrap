# Proposal: Publish Homebrew Stable Channel

## Intent

Publish the already implemented Linux/WSL Homebrew technical slice as a trustworthy public channel. This split corrects the boundary: `homebrew-installation-channel` retains completed resolver/formula approach and 9/9 resolver evidence; this change owns stable publication and channel proof.

**Status: OPEN and BLOCKED** until a real GitHub Release is public, not draft, not prerelease, and includes Linux `amd64`/`arm64` archives with matching SHA-256 assets. A prerelease is validation-only and MUST NOT publish the stable channel.

## Scope

### In Scope
- Qualify a named stable release; record its version, URLs, asset names, and digests with content-level SHA-256 validation.
- Create the pinned tap formula from scratch after the gate passes; prove Linux/WSL install, version, audit/style, reinstall or upgrade, uninstall, no managed-payload residue, and unrelated-file preservation on `amd64` and `arm64`.
- Prove macOS rejection before download; document tap lifecycle, catalog path, platform boundary, and evidence; complete final channel verification.
- Update the main `README.md` with summary install/use instructions only; place detailed publication evidence, hashes, and operational proof in the tap `README.md`.

### Out of Scope
- Source, release, workflow, resolver, archive-layout, or formula-repository changes before the stable gate passes.
- macOS support, formula-update automation, and reworking completed technical evidence.

## Capabilities

### New Capabilities
- `homebrew-stable-channel-publication`: Stable-release qualification, tap publication, lifecycle/platform evidence, and publication blocking.

### Modified Capabilities
- `operational-readme`: Document the supported Linux/WSL Homebrew channel and its stable-release and macOS boundaries.

## Approach

Gate all publication work on release metadata and checksums. Reuse the completed Homebrew resolver/formula contract, pin the verified stable assets, then collect lifecycle and rejection evidence before declaring the channel available.

## Affected Areas

| Area | Impact | Description |
|---|---|---|
| `README.md` | Modified | Homebrew lifecycle and support boundary |
| `dnieblesdev/homebrew-dniebles-bootstrap/` | Modified | Formula, tap guidance, publication evidence |
| `openspec/changes/publish-homebrew-stable-channel/` | New | Acceptance and verification artifacts |

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| No qualifying stable release | High | Remain open/blocked; do not create a release to unblock |
| Wrong asset/digest or lifecycle residue | Medium | Verify release metadata and both architecture runs before publish |

## Rollback Plan

Revert or remove the published formula and tap documentation/evidence; leave the completed resolver and direct/XDG behavior unchanged.

## Dependencies

- Completed `homebrew-installation-channel` technical slice and its 9/9 resolver evidence.
- A qualifying public stable GitHub Release and tap-maintainer access.

## Success Criteria

- [ ] Stable gate evidence satisfies all release and asset conditions before publication.
- [ ] Linux/WSL `amd64` and `arm64` lifecycle evidence, macOS pre-download rejection, and documentation are complete.
- [ ] Final verification confirms no prerelease was published as stable.
