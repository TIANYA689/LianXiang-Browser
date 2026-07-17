package browser

import (
	"encoding/json"
	"lianxiang-browser/backend/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDefaultBookmarksCreatesNestedFolders(t *testing.T) {
	userDataDir := t.TempDir()
	bookmarks := []config.BrowserBookmark{
		{Name: "ChatGPT", URL: "https://chatgpt.com/", Folder: "工作/AI"},
		{Name: "GitHub", URL: "https://github.com/", Folder: "工作/工具"},
		{Name: "Google", URL: "https://www.google.com/"},
	}

	if err := EnsureDefaultBookmarks(userDataDir, bookmarks); err != nil {
		t.Fatalf("首次写入默认书签失败: %v", err)
	}
	if err := EnsureDefaultBookmarks(userDataDir, bookmarks); err != nil {
		t.Fatalf("重复写入默认书签失败: %v", err)
	}

	root := readBookmarkRoot(t, userDataDir)
	barChildren := extractBarChildren(root)
	if got := countBookmarkURL(barChildren, "https://chatgpt.com/"); got != 1 {
		t.Fatalf("ChatGPT 书签数量 = %d，期望 1", got)
	}
	if !folderContainsBookmark(barChildren, []string{"工作", "AI"}, "https://chatgpt.com/") {
		t.Fatal("未在 工作/AI 分组中找到 ChatGPT 书签")
	}
	if !folderContainsBookmark(barChildren, []string{"工作", "工具"}, "https://github.com/") {
		t.Fatal("未在 工作/工具 分组中找到 GitHub 书签")
	}
	if !directChildrenContainURL(barChildren, "https://www.google.com/") {
		t.Fatal("未分组书签没有写入书签栏根目录")
	}
}

func TestEnsureDefaultBookmarksDoesNotDuplicateExistingURLInOtherRoot(t *testing.T) {
	userDataDir := t.TempDir()
	profileDir := filepath.Join(userDataDir, "Default")
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		t.Fatal(err)
	}

	now := toChromiumTime(chromiumEpoch)
	root := newEmptyBookmarkRoot(now)
	roots := root["roots"].(map[string]interface{})
	other := roots["other"].(map[string]interface{})
	other["children"] = []interface{}{map[string]interface{}{
		"id": "4", "name": "Existing", "type": "url", "url": "https://example.com/",
	}}
	data, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(profileDir, "Bookmarks"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := EnsureDefaultBookmarks(userDataDir, []config.BrowserBookmark{
		{Name: "Duplicate", URL: "https://example.com/", Folder: "导入"},
	}); err != nil {
		t.Fatal(err)
	}

	root = readBookmarkRoot(t, userDataDir)
	if got := countRootBookmarkURL(root, "https://example.com/"); got != 1 {
		t.Fatalf("重复 URL 数量 = %d，期望 1", got)
	}
}

func readBookmarkRoot(t *testing.T, userDataDir string) map[string]interface{} {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(userDataDir, "Default", "Bookmarks"))
	if err != nil {
		t.Fatal(err)
	}
	var root map[string]interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	return root
}

func folderContainsBookmark(children []interface{}, folders []string, url string) bool {
	if len(folders) == 0 {
		return directChildrenContainURL(children, url)
	}
	for _, raw := range children {
		folder, ok := raw.(map[string]interface{})
		if !ok || folder["type"] != "folder" || folder["name"] != folders[0] {
			continue
		}
		subChildren, _ := folder["children"].([]interface{})
		return folderContainsBookmark(subChildren, folders[1:], url)
	}
	return false
}

func directChildrenContainURL(children []interface{}, url string) bool {
	for _, raw := range children {
		item, ok := raw.(map[string]interface{})
		if ok && item["type"] == "url" && item["url"] == url {
			return true
		}
	}
	return false
}

func countBookmarkURL(children []interface{}, url string) int {
	total := 0
	for _, raw := range children {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if item["type"] == "url" && item["url"] == url {
			total++
		}
		if subChildren, ok := item["children"].([]interface{}); ok {
			total += countBookmarkURL(subChildren, url)
		}
	}
	return total
}

func countRootBookmarkURL(root map[string]interface{}, url string) int {
	total := 0
	roots, _ := root["roots"].(map[string]interface{})
	for _, raw := range roots {
		folder, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		children, _ := folder["children"].([]interface{})
		total += countBookmarkURL(children, url)
	}
	return total
}
