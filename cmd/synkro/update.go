package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
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
		if err := selfUpdateRun(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func selfUpdateRun(_ []string) error {
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
		return buildFromSource()
	}

	fmt.Printf("📥 Downloading update: %s\n", info.LatestVersion)

	tempFile, err := downloadUpdate(assetURL)
	if err != nil {
		fmt.Printf("⚠️  Error downloading update: %v\n", err)
		return err
	}
	defer os.Remove(tempFile)

	expectedSHA256, err := findChecksumForAsset(info.HTMLURL, platformInfo.binaryName)
	if err == nil && expectedSHA256 != "" {
		fmt.Println("🔍 Verifying checksum...")
		actualSHA256, err := fileSHA256(tempFile)
		if err != nil {
			fmt.Printf("⚠️  Error computing checksum: %v\n", err)
			return err
		}
		if !strings.EqualFold(actualSHA256, expectedSHA256) {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedSHA256, actualSHA256)
		}
		fmt.Println("✅ Checksum verified")
	} else {
		fmt.Println("⚠️  No checksum found, skipping verification")
	}

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
	var newBinary string
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

func buildFromSource() error {
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

func compareVersions(v1, v2 string) int {
	p1 := parseSemver(v1)
	p2 := parseSemver(v2)

	for i := 0; i < 3; i++ {
		if p1[i] > p2[i] {
			return 1
		} else if p1[i] < p2[i] {
			return -1
		}
	}
	return 0
}

func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, "-", 2)
	numbers := strings.Split(parts[0], ".")

	result := [3]int{}
	for i := 0; i < 3 && i < len(numbers); i++ {
		result[i], _ = strconv.Atoi(numbers[i])
	}
	return result
}

func fileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func findChecksumForAsset(releaseURL, assetName string) (string, error) {
	repoURL := strings.TrimSuffix(releaseURL, "/tag/")
	checksumURL := repoURL + "/download/checksums.txt"

	resp, err := http.Get(checksumURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksums file not found (status %d)", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			if strings.HasSuffix(parts[1], assetName) {
				return parts[0], nil
			}
		}
	}

	return "", fmt.Errorf("no checksum found for %s", assetName)
}
