package browser

import (
	"lianxiang-browser/backend/internal/config"
	"lianxiang-browser/backend/internal/database"
	"path/filepath"
	"reflect"
	"testing"
)

func TestSQLiteBookmarkDAOPreservesFolder(t *testing.T) {
	db, err := database.NewDB(filepath.Join(t.TempDir(), "bookmarks.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Migrate(); err != nil {
		t.Fatal(err)
	}

	dao := NewSQLiteBookmarkDAO(db.GetConn())
	want := []config.BrowserBookmark{
		{Name: "GitHub", URL: "https://github.com/", Folder: "工作/工具", OpenOnStart: true, Disabled: true, DisabledProfileIDs: []string{"profile-a"}},
		{Name: "Google", URL: "https://www.google.com/"},
	}
	if err := dao.ReplaceAll(want); err != nil {
		t.Fatal(err)
	}

	got, err := dao.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("书签数量 = %d，期望 %d", len(got), len(want))
	}
	if !reflect.DeepEqual(got[0], want[0]) {
		t.Fatalf("分组书签 = %#v，期望 %#v", got[0], want[0])
	}
}
