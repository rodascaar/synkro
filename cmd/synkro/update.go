package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Name    string `json:"name"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Check for updates",
	Long:  "Check GitHub releases for updates",
	Run: func(cmd *cobra.Command, args []string) {
		latest, err := checkLatestRelease()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking for updates: %v\n", err)
			os.Exit(1)
		}

		if latest.TagName == fmt.Sprintf("v%s", Version) {
			fmt.Println("✅ You are using the latest version!")
			return
		}

		fmt.Printf("🔄 Update available: %s (current: %s)\n", latest.TagName, Version)
		fmt.Printf("📦 Release: %s\n", latest.HTMLURL)
		fmt.Printf("\nRun: curl -fsSL https://raw.githubusercontent.com/rodascaar/synkro/main/install.sh | sh\n")
	},
}

func checkLatestRelease() (*GitHubRelease, error) {
	url := "https://api.github.com/repos/rodascaar/synkro/releases/latest"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
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
