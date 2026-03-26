package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	GitHubRepo       = "VioletEvergar-den/mitm-openai-server"
	GitHubAPIURL     = "https://api.github.com/repos/" + GitHubRepo
	GitHubReleasesURL = "https://github.com/" + GitHubRepo + "/releases"
)

type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Assets      []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	HasUpdate      bool
}

func GetLatestRelease() (*GitHubRelease, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	
	resp, err := client.Get(GitHubAPIURL + "/releases/latest")
	if err != nil {
		return nil, fmt.Errorf("获取最新版本失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取最新版本失败: HTTP %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("解析版本信息失败: %v", err)
	}

	return &release, nil
}

func CompareVersions(current, latest string) int {
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")
	
	maxLen := len(currentParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}
	
	for i := 0; i < maxLen; i++ {
		var c, l int
		if i < len(currentParts) {
			fmt.Sscanf(currentParts[i], "%d", &c)
		}
		if i < len(latestParts) {
			fmt.Sscanf(latestParts[i], "%d", &l)
		}
		
		if c < l {
			return -1
		} else if c > l {
			return 1
		}
	}
	
	return 0
}

func GetBinaryName() string {
	switch runtime.GOOS {
	case "windows":
		return "mitm-openai-server.exe"
	default:
		return "mitm-openai-server"
	}
}

func GetAssetName() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	
	switch osName {
	case "windows":
		return fmt.Sprintf("mitm-openai-server-windows-%s.exe", arch)
	case "darwin":
		return fmt.Sprintf("mitm-openai-server-darwin-%s", arch)
	case "linux":
		return fmt.Sprintf("mitm-openai-server-linux-%s", arch)
	default:
		return fmt.Sprintf("mitm-openai-server-%s-%s", osName, arch)
	}
}

func DownloadFile(url, filepath string) error {
	client := &http.Client{Timeout: 5 * time.Minute}
	
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func GetExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取可执行文件路径失败: %v", err)
	}
	return execPath, nil
}

func CheckForUpdate(currentVersion string) (bool, string, error) {
	release, err := GetLatestRelease()
	if err != nil {
		return false, "", err
	}

	latestVersion := release.TagName
	
	if CompareVersions(currentVersion, latestVersion) < 0 {
		return true, latestVersion, nil
	}
	
	return false, latestVersion, nil
}

func PerformUpdate(currentVersion string) error {
	release, err := GetLatestRelease()
	if err != nil {
		return err
	}

	assetName := GetAssetName()
	var downloadURL string
	
	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	
	if downloadURL == "" {
		return fmt.Errorf("未找到适合 %s/%s 的更新包，请手动下载: %s", runtime.GOOS, runtime.GOARCH, GitHubReleasesURL)
	}

	execPath, err := GetExecutablePath()
	if err != nil {
		return err
	}

	backupPath := execPath + ".backup"
	newPath := execPath + ".new"

	fmt.Printf("正在下载最新版本 %s...\n", release.TagName)
	fmt.Printf("下载地址: %s\n", downloadURL)

	if err := DownloadFile(downloadURL, newPath); err != nil {
		return fmt.Errorf("下载更新失败: %v", err)
	}

	if err := os.Chmod(newPath, 0755); err != nil {
		return fmt.Errorf("设置执行权限失败: %v", err)
	}

	if _, err := os.Stat(backupPath); err == nil {
		os.Remove(backupPath)
	}

	if err := os.Rename(execPath, backupPath); err != nil {
		return fmt.Errorf("备份旧版本失败: %v", err)
	}

	if err := os.Rename(newPath, execPath); err != nil {
		os.Rename(backupPath, execPath)
		return fmt.Errorf("替换可执行文件失败: %v", err)
	}

	os.Remove(backupPath)

	fmt.Printf("\n更新成功! 已更新到版本 %s\n", release.TagName)
	fmt.Println("请重新启动服务器以使用新版本。")
	
	return nil
}

func UpdateViaGit() error {
	fmt.Println("正在通过 Git 拉取最新代码...")
	
	cmd := exec.Command("git", "pull", "origin", "main")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Git 拉取失败: %v", err)
	}
	
	fmt.Println("\n正在重新编译...")
	
	buildCmd := exec.Command("go", "build", "-o", GetBinaryName(), ".")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("编译失败: %v", err)
	}
	
	fmt.Println("\n更新成功! 请重新启动服务器。")
	return nil
}
