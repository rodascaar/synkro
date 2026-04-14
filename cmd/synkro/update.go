package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type GitHubRelease struct {
	TagName    string `json:"tag_name"`
	HTMLURL    string `json:"html_url"`
	Name       string `json:"name"`
	Draft      bool   `json:"draft"`
	PreRelease bool   `json:"prerelease"`
	Body       string `json:"body"`
	Assets     []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
		Size int64  `json:"size"`
	} `json:"assets"`
}

type UpdateInfo struct {
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	HTMLURL         string
	ReleaseNotes    string
	IsPrerelease    bool
	PlatformAssets  map[string]string
}

func checkLatestRelease() (*GitHubRelease, error) {
	resp, err := http.Get("https://api.github.com/repos/rodascaar/synkro/releases/latest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

var checkUpdateCmd = &cobra.Command{
	Use:   "check-update",
	Short: "Check for updates silently",
	Run: func(cmd *cobra.Command, args []string) {
		info, err := getUpdateInfo()
		if err != nil {
			return
		}

		if !info.UpdateAvailable {
			return
		}

		fmt.Printf("🔄 Update available: %s\n", info.LatestVersion)
		fmt.Printf("📦 Release: %s\n", info.HTMLURL)
	},
}

var selfUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Synkro to latest version",
	Long:  "Download and install the latest version of Synkro",
	Run: func(cmd *cobra.Command, args []string) {
		if err := selfUpdateRun(cmd, args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func selfUpdateRun(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 Checking for updates...")

	info, err := getUpdateInfo()
	if err != nil {
		fmt.Printf("⚠️  Error checking for updates: %v\n", err)
		return err
	}

	if !info.UpdateAvailable {
		fmt.Println("✅ Synkro is up to date!")
		return nil
	}

	if info.IsPrerelease {
		fmt.Println("⚠️  Latest version is a pre-release")
		resp := prompt("Continue with update? (y/N): ", false)
		if strings.ToLower(resp) != "y" {
			fmt.Println("Update cancelled")
			return nil
		}
	}

	platformInfo := getPlatform()
	assetURL, ok := info.PlatformAssets[platformInfo.binaryName]

	if !ok {
		fmt.Printf("⚠️  No pre-built binary available for %s\n", platformInfo.binaryName)
		resp := prompt("Build from source instead? (y/N): ", false)
		if strings.ToLower(resp) != "y" {
			fmt.Println("Update cancelled")
			return nil
		}
		return buildFromSource(info.LatestVersion)
	}

	fmt.Printf("📥 Downloading update: %s\n", info.LatestVersion)

	tempFile, err := downloadUpdate(assetURL)
	if err != nil {
		fmt.Printf("⚠️  Error downloading update: %v\n", err)
		return err
	}
	defer os.Remove(tempFile)

	fmt.Println("✅ Download complete")
	fmt.Println("📦 Installing update...")

	err = installUpdate(tempFile, platformInfo)
	if err != nil {
		fmt.Printf("⚠️  Error installing update: %v\n", err)
		return err
	}

	fmt.Println("✅ Update installed successfully!")
	fmt.Printf("🎉 Synkro updated from %s to %s\n", Version, info.LatestVersion)
	fmt.Println("\nPlease restart your terminal to use the new version")

	return nil
}

func getUpdateInfo() (*UpdateInfo, error) {
	release, err := checkLatestRelease()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}

	currentVersion := strings.TrimPrefix(Version, "v")
	latestVersion := strings.TrimPrefix(release.TagName, "v")
	updateAvailable := compareVersions(latestVersion, currentVersion) > 0

	platformAssets := make(map[string]string)
	for _, asset := range release.Assets {
		binaryName := strings.ToLower(asset.Name)
		if strings.HasSuffix(binaryName, ".tar.gz") {
			asset.URL = strings.TrimSuffix(asset.URL, ".tar.gz")
		}
		platformAssets[binaryName] = asset.URL
	}

	return &UpdateInfo{
		CurrentVersion:  currentVersion,
		LatestVersion:   latestVersion,
		UpdateAvailable: updateAvailable,
		HTMLURL:         release.HTMLURL,
		ReleaseNotes:    release.Body,
		IsPrerelease:    release.Draft || release.PreRelease,
		PlatformAssets:  platformAssets,
	}, nil
}

func downloadUpdate(downloadURL string) (string, error) {
	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("synkro-update-%d", time.Now().Unix()))

	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		os.Remove(tempFile)
		return "", fmt.Errorf("failed to save download: %w", err)
	}

	return tempFile, nil
}

func installUpdate(tempFile string, platformInfo platformInfoStruct) error {
	newBinary := tempFile + "-new"
	extractedDir := filepath.Join(os.TempDir(), "synkro-extract")

	switch platformInfo.os {
	case "windows":
		if err := extractTarGz(tempFile, extractedDir); err != nil {
			return fmt.Errorf("failed to extract: %w", err)
		}
		newBinary = filepath.Join(extractedDir, "synkro.exe")

	case "linux", "darwin":
		if err := extractTarGz(tempFile, extractedDir); err != nil {
			return fmt.Errorf("failed to extract: %w", err)
		}
		newBinary = filepath.Join(extractedDir, "synkro")

	default:
		return fmt.Errorf("unsupported platform: %s", platformInfo.os)
	}

	binaryPath := getSynkroPath()

	dir := filepath.Dir(binaryPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	if err := os.Rename(newBinary, binaryPath); err != nil {
		return fmt.Errorf("failed to install: %w", err)
	}

	os.RemoveAll(extractedDir)
	return nil
}

func buildFromSource(version string) error {
	fmt.Println("🔨 Building from source...")

	buildDir := "build"
	os.MkdirAll(buildDir, 0755)

	cmd := exec.Command("go", "build", "-o", "build/synkro", "./cmd/synkro/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	binaryPath := getSynkroPath()
	dir := filepath.Dir(binaryPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory: %w", err)
	}

	if err := os.Rename(filepath.Join(buildDir, "synkro"), binaryPath); err != nil {
		return fmt.Errorf("failed to install: %w", err)
	}

	return nil
}

func extractTarGz(source, dest string) error {
	os.MkdirAll(dest, 0755)

	if runtime.GOOS == "windows" {
		return execCommand("tar", "-xzf", source, "-C", dest)
	}

	return execCommand("tar", "xzf", source, "-C", dest)
}

func execCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getSynkroPath() string {
	if path := os.Getenv("SYNKRO_INSTALL_DIR"); path != "" {
		return filepath.Join(path, "bin", "synkro")
	}

	if runtime.GOOS == "windows" {
		programFiles := os.Getenv("ProgramFiles")
		if programFiles != "" {
			return filepath.Join(programFiles, "synkro", "synkro.exe")
		}
		return filepath.Join(os.Getenv("USERPROFILE"), ".local", "bin", "synkro.exe")
	}

	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".local", "bin", "synkro")
	}

	return "/usr/local/bin/synkro"
}

type platformInfoStruct struct {
	os         string
	arch       string
	binaryName string
}

func getPlatform() platformInfoStruct {
	os := runtime.GOOS
	arch := runtime.GOARCH

	binaryName := "synkro"
	if os == "windows" {
		binaryName += ".exe"
	}

	return platformInfoStruct{
		os:         os,
		arch:       arch,
		binaryName: binaryName,
	}
}

func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		n1, _ := parseVersionPart(parts1[i])
		n2, _ := parseVersionPart(parts2[i])

		if n1 > n2 {
			return 1
		} else if n1 < n2 {
			return -1
		}
	}

	return 0
}

func parseVersionPart(s string) (int, error) {
	s = strings.TrimPrefix(s, "v")
	parts := strings.Split(s, "-")

	if len(parts) == 3 {
		major, _ := strconv.Atoi(parts[0])
		minor, _ := strconv.Atoi(parts[1])
		patch, _ := strconv.Atoi(parts[2])

		return major*10000 + minor*100 + patch, nil
	}

	n, err := strconv.Atoi(s)
	return n, err
}

func prompt(question string, required bool) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(question)

	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)

	if required && answer == "" {
		return prompt(question, required)
	}

	return answer
}
