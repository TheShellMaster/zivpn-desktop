package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

const CurrentVersion = "v1.0.1"
const RepoReleasesAPI = "https://api.github.com/repos/TheShellMaster/zivpn-desktop/releases/latest"
const RepoReleasesURL = "https://github.com/TheShellMaster/zivpn-desktop/releases/latest"

type githubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
}

func CheckForUpdates(w fyne.Window, silentIfLatest bool) {
	go func() {
		client := http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequest("GET", RepoReleasesAPI, nil)
		if err != nil {
			return
		}
		req.Header.Set("User-Agent", "ZiVPN-Desktop-App")

		resp, err := client.Do(req)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return
		}

		var rel githubRelease
		if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
			return
		}

		latest := strings.TrimSpace(rel.TagName)
		if latest != "" && isNewerVersion(latest, CurrentVersion) {
			targetURL := rel.HTMLURL
			if targetURL == "" {
				targetURL = RepoReleasesURL
			}

			// Affichage sur le thread UI
			dialog.ShowConfirm(
				"Mise à jour disponible ! 🎉",
				fmt.Sprintf("Une nouvelle version (%s) de ZiVPN Desktop est disponible.\n\nSouhaitez-vous vous rendre sur la page de téléchargement ?", latest),
				func(confirm bool) {
					if confirm {
						parsedURL, err := url.Parse(targetURL)
						if err == nil {
							_ = fyne.CurrentApp().OpenURL(parsedURL)
						}
					}
				},
				w,
			)
		} else if !silentIfLatest {
			dialog.ShowInformation("À jour", "Vous disposez déjà de la dernière version de ZiVPN Desktop.", w)
		}
	}()
}

// isNewerVersion compare simplement deux tags (ex: "v1.0.1" et "v1.0.0")
func isNewerVersion(latest, current string) bool {
	l := strings.TrimPrefix(latest, "v")
	c := strings.TrimPrefix(current, "v")
	return l != c && l > c
}
