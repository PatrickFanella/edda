package srd521

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func loadRegister(t *testing.T) Register {
	t.Helper()
	r, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestBundledRegister(t *testing.T) {
	r := loadRegister(t)
	if err := r.Validate(os.DirFS("../../..")); err != nil {
		t.Fatal(err)
	}
	// This list protects the original stable IDs from accidental deletion or
	// renaming. New, narrower content IDs may be added by subsequent slices.
	required := map[string]string{}
	for _, family := range strings.Fields("d20-tests actions combat conditions exploration toolbox character-creation backgrounds species feats equipment spells magic-items monsters animals") {
		required["srd-5.2.1:family/"+family] = family
	}
	for class, subclass := range map[string]string{
		"barbarian": "path-of-the-berserker", "bard": "college-of-lore", "cleric": "life-domain",
		"druid": "circle-of-the-land", "fighter": "champion", "monk": "warrior-of-the-open-hand",
		"paladin": "oath-of-devotion", "ranger": "hunter", "rogue": "thief",
		"sorcerer": "draconic-sorcery", "warlock": "fiend-patron", "wizard": "evoker",
	} {
		required["srd-5.2.1:class/"+class] = "classes"
		required["srd-5.2.1:subclass/"+class+"/"+subclass] = "subclasses"
	}
	for level := 1; level <= 20; level++ {
		required[fmt.Sprintf("srd-5.2.1:level/%d", level)] = "advancement"
	}
	for _, e := range r.Entries {
		if family, ok := required[e.ID]; ok {
			if e.Family != family {
				t.Errorf("%s: family = %q, want %q", e.ID, e.Family, family)
			}
			delete(required, e.ID)
		}
	}
	for id := range required {
		t.Errorf("missing stable coverage ID %s", id)
	}
}

func TestImplementationClaimsRequireReviewedEvidence(t *testing.T) {
	// No scope has been qualified against SRD 5.2.1 yet. A later slice must
	// deliberately add its ID and evidence here, alongside real mechanical
	// tests, rather than marking a whole family supported because a file exists.
	qualified := map[string][]string{}
	for _, e := range loadRegister(t).Entries {
		fixtures, ok := qualified[e.ID]
		if e.Status == "implemented" && !ok {
			t.Errorf("%s: unqualified full implementation claim", e.ID)
		}
		if ok && (e.Status != "implemented" || strings.Join(e.Fixtures, "\n") != strings.Join(fixtures, "\n")) {
			t.Errorf("%s: status/fixtures differ from reviewed evidence", e.ID)
		}
		delete(qualified, e.ID)
	}
	for id := range qualified {
		t.Errorf("reviewed evidence refers to missing entry %s", id)
	}
}

func TestValidateRejectsInvalidRegister(t *testing.T) {
	cases := []struct {
		name string
		edit func(*Register)
		want string
	}{
		{"schema", func(r *Register) { r.SchemaVersion = 2 }, "schema"},
		{"version", func(r *Register) { r.Source.Version = "5.1" }, "pinned"},
		{"language", func(r *Register) { r.Source.Language = "fr" }, "pinned"},
		{"artifact URL", func(r *Register) { r.Source.ArtifactURL = "https://example.com/srd.pdf" }, "pinned"},
		{"landing URL", func(r *Register) { r.Source.LandingURL = "https://example.com" }, "pinned"},
		{"hash", func(r *Register) { r.Source.SHA256 = strings.Repeat("0", 64) }, "pinned"},
		{"bytes", func(r *Register) { r.Source.Bytes++ }, "pinned"},
		{"pages", func(r *Register) { r.Source.Pages++ }, "pinned"},
		{"license", func(r *Register) { r.Source.License = "OGL" }, "license"},
		{"license URL", func(r *Register) { r.Source.LicenseURL = "" }, "license"},
		{"attribution", func(r *Register) { r.Source.Attribution = "abbreviated" }, "attribution"},
		{"changes", func(r *Register) { r.Source.Changes = "" }, "change notice"},
		{"date", func(r *Register) { r.Source.RetrievedAt = "yesterday" }, "date"},
		{"empty register", func(r *Register) { r.Entries = nil }, "entries"},
		{"duplicate ID", func(r *Register) { r.Entries = append(r.Entries, r.Entries[0]) }, "duplicate"},
		{"invalid ID", func(r *Register) { r.Entries[0].ID = "d20-tests" }, "content ID"},
		{"wrong version ID", func(r *Register) { r.Entries[0].ID = "srd-5.2:family/d20-tests" }, "content ID"},
		{"invalid level", func(r *Register) { r.Entries[0].ID = "srd-5.2.1:level/21" }, "content ID"},
		{"non-numeric level", func(r *Register) { r.Entries[0].ID = "srd-5.2.1:level/one" }, "content ID"},
		{"subclass without class", func(r *Register) { r.Entries[0].ID = "srd-5.2.1:subclass/champion" }, "content ID"},
		{"content without family", func(r *Register) { r.Entries[0].ID = "srd-5.2.1:content/fireball" }, "content ID"},
		{"family", func(r *Register) { r.Entries[0].Family = "" }, "family"},
		{"name", func(r *Register) { r.Entries[0].Name = " " }, "name"},
		{"notes", func(r *Register) { r.Entries[0].Notes = "" }, "notes"},
		{"follow-up", func(r *Register) { r.Entries[0].FollowUp = -1 }, "follow-up"},
		{"missing range", func(r *Register) { r.Entries[0].SourcePages = nil }, "page range"},
		{"long range", func(r *Register) { r.Entries[0].SourcePages = []int{1, 2, 3} }, "page range"},
		{"zero page", func(r *Register) { r.Entries[0].SourcePages = []int{0, 8} }, "page range"},
		{"reversed pages", func(r *Register) { r.Entries[0].SourcePages = []int{8, 6} }, "page range"},
		{"page overflow", func(r *Register) { r.Entries[0].SourcePages = []int{6, 365} }, "page range"},
		{"unknown status", func(r *Register) { r.Entries[0].Status = "fully-implemented" }, "status"},
		{"implemented without fixtures", func(r *Register) {
			r.Entries[0].Status = "implemented"
			r.Entries[0].Fixtures = []string{}
		}, "requires conformance fixtures"},
		{"null fixtures", func(r *Register) { r.Entries[0].Fixtures = nil }, "explicit array"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := loadRegister(t)
			tc.edit(&r)
			if err := r.Validate(os.DirFS("../../..")); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tc.want)
			}
		})
	}
}

func TestFixtureLinks(t *testing.T) {
	repo := fstest.MapFS{
		"internal/rules/action_test.go":       {Data: []byte("package rules\n")},
		"internal/rules/testdata/action.json": {Data: []byte(`{"case":"example"}`)},
		"internal/rules/testdata/empty.json":  {Data: nil},
		"internal/rules/testdata/directory":   {Mode: os.ModeDir},
		"docs/notes.md":                       {Data: []byte("not a fixture")},
	}
	for _, tc := range []struct {
		name  string
		paths []string
		valid bool
	}{
		{"Go test", []string{"internal/rules/action_test.go"}, true},
		{"data fixture", []string{"internal/rules/testdata/action.json"}, true},
		{"missing", []string{"internal/rules/testdata/missing.json"}, false},
		{"absolute", []string{"/tmp/action_test.go"}, false},
		{"traversal", []string{"../action_test.go"}, false},
		{"unclean", []string{"internal/../action_test.go"}, false},
		{"backslash", []string{`internal\rules\action_test.go`}, false},
		{"empty path", []string{""}, false},
		{"root", []string{"."}, false},
		{"duplicate", []string{"internal/rules/action_test.go", "internal/rules/action_test.go"}, false},
		{"directory", []string{"internal/rules/testdata/directory"}, false},
		{"empty file", []string{"internal/rules/testdata/empty.json"}, false},
		{"prose", []string{"docs/notes.md"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := loadRegister(t)
			r.Entries[0].Status = "implemented"
			r.Entries[0].Fixtures = tc.paths
			if err := r.Validate(repo); (err == nil) != tc.valid {
				t.Fatalf("Validate() = %v, valid = %v", err, tc.valid)
			}
			if tc.valid && r.Validate(nil) == nil {
				t.Fatal("accepted fixtures without a repository filesystem")
			}
		})
	}
}

func TestLoadReturnsIndependentCopy(t *testing.T) {
	r := loadRegister(t)
	r.Entries[0].SourcePages[0] = 999
	r.Entries[0].Status = "implemented"
	if err := loadRegister(t).Validate(os.DirFS("../../..")); err != nil {
		t.Fatalf("mutating a loaded copy changed the bundled register: %v", err)
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestVerifyArtifactRejectsMismatch(t *testing.T) {
	for _, tc := range []struct {
		name string
		pdf  io.Reader
	}{
		{"empty", strings.NewReader("")},
		{"wrong file", strings.NewReader("not the SRD")},
		{"same size wrong hash", bytes.NewReader(bytes.Repeat([]byte{'x'}, int(artifactBytes)))},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := VerifyArtifact(tc.pdf); err == nil {
				t.Fatal("accepted an artifact that does not match the pin")
			}
		})
	}
	readErr := errors.New("read failed")
	if err := VerifyArtifact(failingReader{readErr}); !errors.Is(err, readErr) {
		t.Fatalf("VerifyArtifact() = %v, want wrapped read failure", err)
	}
}

func TestPinnedArtifact(t *testing.T) {
	// Opt-in, offline proof against the actual PDF; normal CI needs no download.
	name := os.Getenv("EDDA_SRD_PDF")
	if name == "" {
		t.Skip("set EDDA_SRD_PDF to verify a local official PDF")
	}
	f, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := VerifyArtifact(f); err != nil {
		t.Fatal(err)
	}
}
