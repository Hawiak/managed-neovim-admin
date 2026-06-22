# managed-neovim-admin

Admin tooling for organisations deploying managed-nvim. Use this repo to onboard a new org, manage the approved plugin list, and mirror plugins to your registry.

---

## Getting started (new organisation)

### Prerequisites

- Go 1.21+
- Git
- Access to an Artifactory instance (generic repository)
- A GitHub org where you can create repos and set secrets

### Step 1 — Build the tools

```bash
git clone https://github.com/Hawiak/managed-neovim-admin
cd managed-neovim-admin
go build ./cmd/nvim-manifest/
go build ./cmd/nvim-mirror/
go build ./cmd/nvim-distro-init/
```

### Step 2 — Generate a signing key pair

```bash
./nvim-manifest keygen
```

Output:
```
Private key written to: signing.key
Store it in CI secrets as MANIFEST_SIGNING_KEY — never commit it.

Public key (paste as --key flag and MANIFEST_PUBLIC_KEY secret):
ABC123...base64...
```

Keep `signing.key` secret. You will need it in Step 5. The base64 public key goes into Steps 3 and 6.

### Step 3 — Create the distro repo

```bash
./nvim-distro-init \
  --org mycompany \
  --url https://artifactory.mycompany.com/managed-neovim \
  --key <base64-public-key-from-step-2>
```

This creates `mycompany-nvim-distro/` containing:
- `install.sh` — the employee install script
- `managed-neovim.toml` — wrapper config with your URL and signing key baked in
- `manifest/plugins.json` — empty starter manifest
- `.github/workflows/release.yml` — pipeline that builds and publishes the wrapper binary

### Step 4 — Push the distro repo to GitHub

```bash
cd mycompany-nvim-distro
git remote add origin https://github.com/mycompany/mycompany-nvim-distro
git push -u origin main
```

### Step 5 — Add CI secrets to the distro repo

In the GitHub repo settings → Secrets and variables → Actions, add:

| Secret | Value |
|---|---|
| `MANIFEST_PUBLIC_KEY` | base64 public key from Step 2 |
| `ARTIFACTORY_TOKEN` | Artifactory token with write access |

### Step 6 — Add plugins to the manifest

```bash
cd managed-neovim-admin
./nvim-manifest add folke/lazy.nvim <your-name>
./nvim-manifest add nvim-treesitter/nvim-treesitter <your-name>
# add the rest of the org's approved plugins
```

### Step 7 — Mirror plugins to Artifactory

```bash
GITEA_URL=https://artifactory.mycompany.com \
GITEA_OWNER=mycompany \
GITEA_TOKEN=<artifactory-token> \
./nvim-mirror
```

### Step 8 — Sign the manifest

```bash
./nvim-manifest sign signing.key
```

Produces `manifest/plugins.json.sig`. Commit and push both files.

### Step 9 — Tag a release to publish the wrapper binary

```bash
cd mycompany-nvim-distro
git tag v1.0.0 && git push --tags
```

This triggers `.github/workflows/release.yml` which builds `nvim-wrapper` for all four platforms (macOS/Linux × arm64/amd64) and publishes the binaries + checksums to Artifactory.

### Step 10 — Share the install one-liner with employees

```bash
curl -fsSL https://raw.githubusercontent.com/mycompany/mycompany-nvim-distro/main/install.sh | sudo bash
```

---

## Ongoing — adding a plugin

1. Receive a plugin request from a developer
2. Review the plugin source for anything suspicious
3. Add it to the manifest:
   ```bash
   ./nvim-manifest add <owner/repo> <your-name>
   ```
4. Mirror it to Artifactory:
   ```bash
   GITEA_OWNER=mycompany GITEA_TOKEN=<token> ./nvim-mirror
   ```
5. Sign the updated manifest:
   ```bash
   ./nvim-manifest sign signing.key
   ```
6. Commit and push `manifest/plugins.json` and `manifest/plugins.json.sig`
7. The wrapper on employee machines picks up the new manifest on the next `nvim` launch

---

## Ongoing — publishing a new wrapper version

When `managed-nvim` releases a new version:

1. Update `--nvim-ref` in your distro by re-running `nvim-distro-init`, or edit `.github/workflows/release.yml` directly to update the `ref`
2. Tag a new release in the distro repo:
   ```bash
   git tag v1.1.0 && git push --tags
   ```
3. Update `VERSION` in `install.sh` to the new tag so new installs get the latest binary

---

## Tool reference

### nvim-manifest

```bash
nvim-manifest keygen                        # generate a new ed25519 key pair
nvim-manifest add <owner/repo> <approved-by>
nvim-manifest remove <plugin-name>
nvim-manifest verify                        # check manifest is well-formed
nvim-manifest sign <private-key-file>       # CI only — produces plugins.json.sig
nvim-manifest verify-sig <public-key-file>  # verify the signature
```

Manifest path defaults to `manifest/plugins.json`. Override with `MANIFEST_PATH`.

### nvim-mirror

```bash
GITEA_URL=https://...      # default: http://localhost:2222
GITEA_OWNER=myorg          # required
GITEA_TOKEN=<token>        # required
MANIFEST_PATH=...          # default: manifest/plugins.json
./nvim-mirror
```

Clones each plugin at its pinned SHA, tarballs it, and PUT to Artifactory. Exits non-zero if any plugin fails.

### nvim-distro-init

```bash
nvim-distro-init \
  --org <name>          \   # required
  --url <artifactory>   \   # required — Artifactory base URL
  --key <base64-pubkey> \   # required — from nvim-manifest keygen
  --version <version>   \   # default: latest
  --nvim-repo <owner/repo>\ # default: Hawiak/managed-nvim
  --nvim-ref <ref>          # default: main — pin to a tag for stability
```

---

## Local development

A Gitea instance is included for testing the mirror pipeline without hitting production:

```bash
docker compose up -d   # Gitea at http://localhost:2222
GITEA_URL=http://localhost:2222 GITEA_OWNER=test GITEA_TOKEN=<token> ./nvim-mirror
```

---

## Repository layout

```
managed-neovim-admin/
├── cmd/
│   ├── nvim-manifest/
│   │   ├── main.go
│   │   ├── keygen.go
│   │   ├── add.go
│   │   ├── remove.go
│   │   ├── sign.go
│   │   ├── verify.go
│   │   └── verify_sig.go
│   ├── nvim-mirror/
│   │   ├── main.go
│   │   ├── mirror.go
│   │   ├── archive.go
│   │   └── gitea.go
│   └── nvim-distro-init/
│       ├── main.go
│       ├── install.sh.tmpl
│       ├── managed-neovim.toml.tmpl
│       └── release.yml.tmpl
├── internal/manifest/
│   └── manifest.go
├── manifest/
│   └── plugins.json
└── docker-compose.yml
```
