package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ystsbry/bako/internal/model"
)

// setHome points BAKO_HOME at a fresh temp dir for the duration of the test.
func setHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("BAKO_HOME", dir)
	return dir
}

func TestProjectCreateListLoad(t *testing.T) {
	setHome(t)

	p, err := CreateProject("My First Project", "a description\nwith two lines")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if p.Slug != "my-first-project" {
		t.Fatalf("slug = %q, want my-first-project", p.Slug)
	}

	got, err := LoadProject(p.Slug)
	if err != nil {
		t.Fatalf("LoadProject: %v", err)
	}
	if got.Name != "My First Project" {
		t.Errorf("Name = %q", got.Name)
	}
	if got.Description != "a description\nwith two lines" {
		t.Errorf("Description = %q", got.Description)
	}

	list, err := ListProjects()
	if err != nil {
		t.Fatalf("ListProjects: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("len(list) = %d, want 1", len(list))
	}
}

func TestProjectSlugCollision(t *testing.T) {
	setHome(t)
	a, _ := CreateProject("Same Name", "")
	b, _ := CreateProject("Same Name", "")
	if a.Slug == b.Slug {
		t.Fatalf("expected unique slugs, both %q", a.Slug)
	}
	if b.Slug != "same-name-2" {
		t.Errorf("second slug = %q, want same-name-2", b.Slug)
	}
}

func TestProjectSlugNonASCII(t *testing.T) {
	setHome(t)
	p, err := CreateProject("日本語プロジェクト", "")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if p.Slug != "project" {
		t.Errorf("slug = %q, want fallback 'project'", p.Slug)
	}
	if got, _ := LoadProject(p.Slug); got.Name != "日本語プロジェクト" {
		t.Errorf("Name = %q, want original japanese name", got.Name)
	}
}

func TestPBILifecycle(t *testing.T) {
	setHome(t)
	proj, _ := CreateProject("P", "")

	p1, err := CreatePBI(proj.Slug, "First item", model.StatusTodo, "body one")
	if err != nil {
		t.Fatalf("CreatePBI: %v", err)
	}
	if p1.ID != "001" {
		t.Errorf("first ID = %q, want 001", p1.ID)
	}
	p2, _ := CreatePBI(proj.Slug, "Second item", model.StatusProgress, "body two")
	if p2.ID != "002" {
		t.Errorf("second ID = %q, want 002", p2.ID)
	}

	list, err := ListPBIs(proj.Slug)
	if err != nil {
		t.Fatalf("ListPBIs: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}
	if list[0].Title != "First item" || list[0].Body != "body one" {
		t.Errorf("pbi[0] = %+v", list[0])
	}
	if list[1].Status != model.StatusProgress {
		t.Errorf("pbi[1] status = %q", list[1].Status)
	}

	// Update status and title; ensure only one file remains for the ID.
	p1.Status = model.StatusDone
	p1.Title = "First item renamed"
	if err := SavePBI(proj.Slug, p1); err != nil {
		t.Fatalf("SavePBI: %v", err)
	}
	list, _ = ListPBIs(proj.Slug)
	if len(list) != 2 {
		t.Fatalf("after rename len = %d, want 2 (stale file left behind?)", len(list))
	}
	if list[0].Status != model.StatusDone || list[0].Title != "First item renamed" {
		t.Errorf("after update pbi[0] = %+v", list[0])
	}

	if err := DeletePBI(proj.Slug, "001"); err != nil {
		t.Fatalf("DeletePBI: %v", err)
	}
	list, _ = ListPBIs(proj.Slug)
	if len(list) != 1 || list[0].ID != "002" {
		t.Fatalf("after delete = %+v", list)
	}
}

func TestRepoSaveListRepoAndOrg(t *testing.T) {
	setHome(t)
	proj, _ := CreateProject("P", "")

	repo := model.Repo{Kind: model.RepoKindRepo, Owner: "ystsbry", Name: "bako", Overview: "tracker"}
	if err := SaveRepo(proj.Slug, repo); err != nil {
		t.Fatalf("SaveRepo: %v", err)
	}
	org := model.Repo{Kind: model.RepoKindOrg, Owner: "charmbracelet"}
	if err := SaveRepo(proj.Slug, org); err != nil {
		t.Fatalf("SaveRepo org: %v", err)
	}

	list, err := ListRepos(proj.Slug)
	if err != nil {
		t.Fatalf("ListRepos: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("len = %d, want 2", len(list))
	}

	var gotRepo, gotOrg *model.Repo
	for i := range list {
		switch list[i].Kind {
		case model.RepoKindRepo:
			gotRepo = &list[i]
		case model.RepoKindOrg:
			gotOrg = &list[i]
		}
	}
	if gotRepo == nil || gotRepo.URL != "https://github.com/ystsbry/bako" {
		t.Errorf("repo URL not auto-filled: %+v", gotRepo)
	}
	if gotOrg == nil || gotOrg.URL != "https://github.com/charmbracelet" || gotOrg.Name != "" {
		t.Errorf("org entry wrong: %+v", gotOrg)
	}

	// Re-saving same identity overwrites rather than duplicates.
	repo.Overview = "updated"
	_ = SaveRepo(proj.Slug, repo)
	list, _ = ListRepos(proj.Slug)
	if len(list) != 2 {
		t.Fatalf("after re-save len = %d, want 2", len(list))
	}
}

func TestSaveRepoValidation(t *testing.T) {
	setHome(t)
	proj, _ := CreateProject("P", "")
	if err := SaveRepo(proj.Slug, model.Repo{Kind: model.RepoKindRepo, Owner: "x"}); err == nil {
		t.Error("expected error for repo without name")
	}
	if err := SaveRepo(proj.Slug, model.Repo{Kind: model.RepoKindRepo, Name: "y"}); err == nil {
		t.Error("expected error for repo without owner")
	}
}

func TestDeleteProject(t *testing.T) {
	dir := setHome(t)
	proj, _ := CreateProject("Gone", "")
	if err := DeleteProject(proj.Slug); err != nil {
		t.Fatalf("DeleteProject: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, proj.Slug)); !os.IsNotExist(err) {
		t.Errorf("project dir still exists")
	}
	if ProjectExists(proj.Slug) {
		t.Errorf("ProjectExists true after delete")
	}
}

func TestFrontmatterRoundTrip(t *testing.T) {
	meta := pbiMeta{Title: "T", Status: "todo", Created: "2026-06-26"}
	data, err := renderDoc(meta, "line1\nline2")
	if err != nil {
		t.Fatalf("renderDoc: %v", err)
	}
	var got pbiMeta
	body, err := parseDoc(data, &got)
	if err != nil {
		t.Fatalf("parseDoc: %v", err)
	}
	if got != meta {
		t.Errorf("meta round-trip: got %+v want %+v", got, meta)
	}
	if body != "line1\nline2\n" {
		t.Errorf("body = %q", body)
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Hello World":    "hello-world",
		"  Foo--Bar  ":   "foo-bar",
		"日本語":            "",
		"v1.2.3 Release": "v1-2-3-release",
	}
	for in, want := range cases {
		if got := Slugify(in); got != want {
			t.Errorf("Slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
