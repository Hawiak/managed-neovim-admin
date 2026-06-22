# managed-neovim-admin

Admin tooling for organisations deploying [managed-nvim](../managed-nvim). Use this repo to manage your plugin manifest, mirror plugins to your registry, and generate employee-facing distros.

---

## nvim-distro-init

Scaffolds an org-specific distro repo. Run this once when onboarding a new organisation.

```bash
nvim-distro-init \
  --org mycompany \
  --url https://artifactory.mycompany.com/managed-neovim \
  --key <base64-ed25519-pubkey>
```

**Flags:**
| Flag | Required | Description |
|---|---|---|
| `--org` | yes | Organisation name — used as the output directory prefix |
| `--url` | yes | Artifactory base URL — baked into `install.sh` and `managed-neovim.toml` |
| `--key` | yes | base64-encoded ed25519 public key — baked into `managed-neovim.toml` |
| `--version` | no | Wrapper version to install (default: `latest`) |

Outputs `mycompany-nvim-distro/` — a ready-to-push git repo containing:
- `install.sh` — hardcoded with the org's Artifactory URL and version
- `managed-neovim.toml` — pre-filled with URL and signing key
- `manifest/plugins.json` — empty starter manifest

Push it to GitHub and share the install one-liner with employees:

```bash
curl -fsSL https://raw.githubusercontent.com/mycompany/mycompany-nvim-distro/main/install.sh | sudo bash
```

---

## nvim-manifest

Manages `manifest/plugins.json` — the approved plugin list. All changes go through this tool so the manifest stays consistent and auditable.

```bash
nvim-manifest add <owner/repo> <approved-by>    # add a plugin
nvim-manifest remove <plugin-name>              # remove a plugin
nvim-manifest verify                            # check all SHAs are valid
nvim-manifest sign <private-key-file>           # CI only — produces plugins.json.sig
nvim-manifest verify-sig <public-key-file>      # verify the signature
```

The manifest path defaults to `manifest/plugins.json`. Override with `MANIFEST_PATH=path/to/plugins.json`.

`sign` and `verify-sig` use ed25519. The private key is PEM-encoded (`PRIVATE KEY` block). Keep it in CI secrets only — never commit it.

---

## nvim-mirror

Syncs every plugin in the manifest to a Gitea generic package registry as a `.tar.gz`. Run in CI after every manifest change.

**Environment variables:**
| Variable | Required | Default | Description |
|---|---|---|---|
| `GITEA_URL` | no | `http://localhost:2222` | Registry base URL |
| `GITEA_OWNER` | yes | — | Gitea user/org that owns the packages |
| `GITEA_TOKEN` | yes | — | API token with package write access |
| `MANIFEST_PATH` | no | `manifest/plugins.json` | Path to plugins.json |

```bash
GITEA_OWNER=myorg GITEA_TOKEN=$TOKEN nvim-mirror
```

For each plugin: clones at the pinned SHA, tarballs the source, and HTTP PUTs it to `<GITEA_URL>/api/packages/<owner>/generic/<repo>/<sha>.tar.gz`. Exits non-zero if any plugin fails.

---

## Typical workflow

1. Developer requests a new plugin via your internal process
2. Admin reviews and runs `nvim-manifest add <owner/repo> <your-name>`
3. CI runs `nvim-mirror` to fetch and publish the plugin archive
4. CI runs `nvim-manifest sign` to publish a new signed manifest
5. Wrapper on employee machines picks up the new manifest on next `nvim` launch

---

## Local development

A Gitea instance is included for testing the full mirror pipeline locally:

```bash
docker compose up -d
```

Gitea runs at `http://localhost:2222`. Set `GITEA_URL=http://localhost:2222` when running `nvim-mirror` locally.

---

## Repository layout

```
managed-neovim-admin/
├── cmd/
│   ├── nvim-manifest/        # manifest management CLI
│   │   ├── main.go
│   │   ├── add.go
│   │   ├── remove.go
│   │   ├── sign.go
│   │   ├── verify.go
│   │   └── verify_sig.go
│   ├── nvim-mirror/          # mirrors plugins to Gitea
│   │   ├── main.go
│   │   ├── mirror.go
│   │   ├── archive.go
│   │   └── gitea.go
│   └── nvim-distro-init/     # scaffolds org distro repos
│       ├── main.go
│       ├── install.sh.tmpl
│       └── managed-neovim.toml.tmpl
├── internal/manifest/
│   └── manifest.go           # Plugin and Manifest types, Load/Save
├── manifest/
│   └── plugins.json          # the org's approved plugin list
└── docker-compose.yml        # local Gitea for testing
```
