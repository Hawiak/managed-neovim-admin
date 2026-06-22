# managed-neovim-admin

Admin tooling for organisations deploying [managed-nvim](../managed-nvim). Use this repo to manage your plugin manifest, mirror plugins to your registry, and generate employee-facing distros.

## Tools

### nvim-distro-init

Scaffolds an org-specific distro repo. Run this once when onboarding a new organisation.

```bash
nvim-distro-init \
  --org mycompany \
  --url https://artifactory.mycompany.com/managed-neovim \
  --key <base64-ed25519-pubkey>
```

Outputs `mycompany-nvim-distro/` — a ready-to-push git repo containing a hardcoded `install.sh`, a `managed-neovim.toml`, and a starter manifest. Push it to GitHub and share the install one-liner with employees:

```bash
curl -fsSL https://raw.githubusercontent.com/mycompany/mycompany-nvim-distro/main/install.sh | sudo bash
```

---

### nvim-manifest

Manages `manifest/plugins.json` — the approved plugin list. All changes go through this tool so the manifest stays consistent and auditable.

```bash
nvim-manifest add <owner/repo> <sha> <approved-by>
nvim-manifest remove <plugin-name>
nvim-manifest verify        # checks all SHAs are valid
nvim-manifest sign <key>    # CI only — produces plugins.json.sig
nvim-manifest verify-sig    # verifies the signature
```

---

### nvim-mirror

Syncs every plugin in the manifest to your package registry as a `.tar.gz`. Run in CI after every manifest change.

```bash
nvim-mirror --registry https://artifactory.mycompany.com --token $REGISTRY_TOKEN
```

---

## Typical workflow

1. Developer requests a new plugin via your internal process
2. Admin reviews and runs `nvim-manifest add <repo> <sha> <your-name>`
3. CI runs `nvim-mirror` to fetch and publish the plugin archive
4. CI runs `nvim-manifest sign` to publish a new signed manifest
5. Wrapper on employee machines picks up the new manifest on next `nvim` launch

## Local development

A Gitea instance is included for testing the full mirror pipeline without hitting a real registry:

```bash
docker compose up -d
```

Gitea runs at `http://localhost:2222`. Set `REGISTRY_URL=http://localhost:2222` when running `nvim-mirror` locally.
