package base

import (
	"context"
	"os"
	"testing"
)

func TestRepo(t *testing.T) {
	urls := []string{
		// ssh://[user@]host.xz[:port]/path/to/repo.git/
		"ssh://git@github.com:7875/ag-core/aggo.git",
		// git://host.xz[:port]/path/to/repo.git/
		"git://github.com:7875/ag-core/aggo.git",
		// http[s]://host.xz[:port]/path/to/repo.git/
		"https://github.com:7875/ag-core/aggo.git",
		// ftp[s]://host.xz[:port]/path/to/repo.git/
		"ftps://github.com:7875/ag-core/aggo.git",
		//[user@]host.xz:path/to/repo.git/
		"git@github.com:ag-core/aggo.git",
		// ssh://[user@]host.xz[:port]/~[user]/path/to/repo.git/
		"ssh://git@github.com:7875/ag-core/aggo.git",
		// git://host.xz[:port]/~[user]/path/to/repo.git/
		"git://github.com:7875/ag-core/aggo.git",
		//[user@]host.xz:/~[user]/path/to/repo.git/
		"git@github.com:ag-core/aggo.git",
		///path/to/repo.git/
		"//github.com/ag-core/aggo.git",
		// file:///path/to/repo.git/
		"file://./github.com/ag-core/aggo.git",
	}
	for _, url := range urls {
		dir := repoDir(url)
		if dir != "github.com/ag-core" && dir != "/ag-core" {
			t.Fatal(url, "repoDir test failed", dir)
		}
	}
}

func TestRepoClone(t *testing.T) {
	r := NewRepo("https://github.com/ZhengweiHou/ag-layout-demo.git", "")
	if err := r.Clone(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := r.CopyTo(context.Background(), "/tmp/test_repo", "github.com/ZhengweiHou/ag-layout-demo", nil); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		os.RemoveAll("/tmp/test_repo")
	})
}
