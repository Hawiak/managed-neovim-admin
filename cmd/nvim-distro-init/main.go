package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

//go:embed install.sh.tmpl managed-neovim.toml.tmpl release.yml.tmpl
var templates embed.FS

type distroConfig struct {
	Org             string
	ArtifactoryURL  string
	SigningKey       string
	Version         string
	ManagedNvimRepo string
	ManagedNvimRef  string
}

func main() {
	org := flag.String("org", "", "Organisation name (used as the output directory name)")
	url := flag.String("url", "", "Artifactory base URL (e.g. https://artifactory.company.com/managed-neovim)")
	key := flag.String("key", "", "ed25519 public key (base64-encoded)")
	version := flag.String("version", "latest", "Wrapper version to install")
	nvimRepo := flag.String("nvim-repo", "Hawiak/managed-nvim", "GitHub repo to build the wrapper from (owner/repo)")
	nvimRef := flag.String("nvim-ref", "main", "Git ref of managed-nvim to build from (tag, branch, or SHA)")
	flag.Parse()

	if *org == "" || *url == "" || *key == "" {
		fmt.Fprintf(os.Stderr, "Usage: nvim-distro-init --org <name> --url <artifactory-url> --key <pubkey>\n")
		os.Exit(1)
	}

	cfg := distroConfig{
		Org:             *org,
		ArtifactoryURL:  *url,
		SigningKey:       *key,
		Version:         *version,
		ManagedNvimRepo: *nvimRepo,
		ManagedNvimRef:  *nvimRef,
	}

	outDir := *org + "-nvim-distro"
	if err := generate(outDir, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "nvim-distro-init: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Distro created at ./%s\n\n", outDir)
	fmt.Printf("Next steps:\n")
	fmt.Printf("  1. cd %s\n", outDir)
	fmt.Printf("  2. Add secrets to the GitHub repo:\n")
	fmt.Printf("       MANIFEST_PUBLIC_KEY — base64 ed25519 public key\n")
	fmt.Printf("       ARTIFACTORY_TOKEN   — token with write access to %s\n", *url)
	fmt.Printf("  3. Push and tag a release:\n")
	fmt.Printf("       git push && git tag v1.0.0 && git push --tags\n")
	fmt.Printf("  4. Share the install one-liner with employees:\n")
	fmt.Printf("       curl -fsSL https://raw.githubusercontent.com/<org>/%s/main/install.sh | sudo bash\n", outDir)
}

func generate(outDir string, cfg distroConfig) error {
	if _, err := os.Stat(outDir); err == nil {
		return fmt.Errorf("output directory %q already exists", outDir)
	}

	dirs := []string{
		outDir,
		filepath.Join(outDir, "manifest"),
		filepath.Join(outDir, ".github", "workflows"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	if err := renderTemplate("install.sh.tmpl", filepath.Join(outDir, "install.sh"), cfg, 0755); err != nil {
		return err
	}

	if err := renderTemplate("managed-neovim.toml.tmpl", filepath.Join(outDir, "managed-neovim.toml"), cfg, 0644); err != nil {
		return err
	}

	if err := renderTemplate("release.yml.tmpl", filepath.Join(outDir, ".github", "workflows", "release.yml"), cfg, 0644); err != nil {
		return err
	}

	if err := writeStarterManifest(outDir, cfg); err != nil {
		return err
	}

	if err := initGitRepo(outDir); err != nil {
		return fmt.Errorf("git init: %w", err)
	}

	return nil
}

func renderTemplate(tmplName, dest string, cfg distroConfig, mode os.FileMode) error {
	raw, err := templates.ReadFile(tmplName)
	if err != nil {
		return err
	}

	tmpl, err := template.New(tmplName).Parse(string(raw))
	if err != nil {
		return err
	}

	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, cfg)
}

func writeStarterManifest(outDir string, cfg distroConfig) error {
	content := fmt.Sprintf(`{
  "schema_version": 1,
  "org_name": %q,
  "last_updated": "",
  "plugins": []
}
`, cfg.Org)
	return os.WriteFile(filepath.Join(outDir, "manifest", "plugins.json"), []byte(content), 0644)
}

func initGitRepo(dir string) error {
	for _, args := range [][]string{
		{"init"},
		{"add", "."},
		{"commit", "-m", "Initial distro scaffold"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}
