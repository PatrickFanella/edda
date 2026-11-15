// Package srd521 records conformance to the pinned English SRD 5.2.1 and
// implements narrow character-sheet fixtures without changing legacy rules.
package srd521

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"regexp"
	"strings"
	"time"
)

// Version identifies the reference, independently of Edda's release version.
const Version = "5.2.1"

// Attribution is the exact statement required by the source's legal notice (page 1).
const Attribution = "This work includes material from the System Reference Document 5.2.1 (“SRD 5.2.1”) by Wizards of the Coast LLC, available at https://www.dndbeyond.com/srd. The SRD 5.2.1 is licensed under the Creative Commons Attribution 4.0 International License, available at https://creativecommons.org/licenses/by/4.0/legalcode."

const artifactSHA256 = "8974902d109d6e63672d7c490bde9ccf052410503d9cfa768237154fbc5e3d87"
const artifactBytes int64 = 6031375

//go:embed coverage.json
var coverageJSON []byte

// Source pins the original artifact and its notice, not a third-party catalogue.
type Source struct {
	Version     string `json:"version"`
	Language    string `json:"language"`
	LandingURL  string `json:"landing_url"`
	ArtifactURL string `json:"artifact_url"`
	SHA256      string `json:"sha256"`
	Bytes       int64  `json:"bytes"`
	Pages       int    `json:"pages"`
	RetrievedAt string `json:"retrieved_at"`
	License     string `json:"license"`
	LicenseURL  string `json:"license_url"`
	Attribution string `json:"attribution"`
	Changes     string `json:"changes"`
}

// Entry identifies a coverage scope. SourcePages is an inclusive, one-based range
// of PDF pages. Fixtures are repository-relative files containing conformance cases.
// Implemented means the entire named scope has evidence; narrower support must
// get its own entry and must not promote its containing family, class or level.
type Entry struct {
	ID          string   `json:"id"`
	Family      string   `json:"family"`
	Name        string   `json:"name"`
	SourcePages []int    `json:"source_pages"`
	Status      string   `json:"status"`
	Fixtures    []string `json:"fixtures"`
	FollowUp    int      `json:"follow_up"`
	Notes       string   `json:"notes"`
}

// Register is the versioned contract shared by subsequent rules slices.
type Register struct {
	SchemaVersion int     `json:"schema_version"`
	Source        Source  `json:"source"`
	Entries       []Entry `json:"entries"`
}

// Load decodes a fresh copy of the bundled register. Validate it with a repository
// filesystem when checking fixture links; no filesystem or network is used here.
func Load() (Register, error) {
	var r Register
	d := json.NewDecoder(bytes.NewReader(coverageJSON))
	d.DisallowUnknownFields()
	if err := d.Decode(&r); err != nil {
		return Register{}, fmt.Errorf("decode SRD register: %w", err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return Register{}, fmt.Errorf("SRD register must contain one JSON object")
	}
	return r, nil
}

const slug = `[a-z0-9]+(-[a-z0-9]+)*`

var contentID = regexp.MustCompile(`^srd-5\.2\.1:((family|class)/` + slug + `|subclass/` + slug + `/` + slug + `|level/([1-9]|1[0-9]|20)|content/` + slug + `/` + slug + `)$`)
var familyName = regexp.MustCompile(`^` + slug + `$`)

// Validate checks the pin, IDs, page bounds, status claims and fixture paths.
// repo must be rooted at the repository (for example os.DirFS("../../..")); a nil
// filesystem rejects fixture links. Passing validation is not proof of rules
// correctness: reviewed tests must substantiate the entire scope before promotion.
func (r Register) Validate(repo fs.FS) error {
	if r.SchemaVersion != 1 {
		return fmt.Errorf("unsupported register schema %d", r.SchemaVersion)
	}
	s := r.Source
	if s.Version != Version || s.Language != "en" ||
		s.LandingURL != "https://www.dndbeyond.com/srd" ||
		s.ArtifactURL != "https://media.dndbeyond.com/compendium-images/srd/5.2/SRD_CC_v5.2.1.pdf" ||
		s.SHA256 != artifactSHA256 || s.Bytes != artifactBytes || s.Pages != 364 {
		return fmt.Errorf("source does not match the pinned English SRD %s artifact", Version)
	}
	if s.License != "CC-BY-4.0" || s.LicenseURL != "https://creativecommons.org/licenses/by/4.0/legalcode" ||
		s.Attribution != Attribution || strings.TrimSpace(s.Changes) == "" {
		return fmt.Errorf("source must retain its license, exact attribution and change notice")
	}
	if _, err := time.Parse(time.DateOnly, s.RetrievedAt); err != nil {
		return fmt.Errorf("invalid source retrieval date: %w", err)
	}
	if len(r.Entries) == 0 {
		return fmt.Errorf("register must contain coverage entries")
	}
	ids := make(map[string]bool, len(r.Entries))
	for _, e := range r.Entries {
		if !contentID.MatchString(e.ID) || ids[e.ID] {
			return fmt.Errorf("invalid or duplicate content ID %q", e.ID)
		}
		ids[e.ID] = true
		if !familyName.MatchString(e.Family) || strings.TrimSpace(e.Name) == "" || strings.TrimSpace(e.Notes) == "" || e.FollowUp < 0 {
			return fmt.Errorf("%s: missing or invalid family, name, notes or follow-up", e.ID)
		}
		if len(e.SourcePages) != 2 || e.SourcePages[0] < 1 || e.SourcePages[0] > e.SourcePages[1] || e.SourcePages[1] > s.Pages {
			return fmt.Errorf("%s: invalid inclusive source page range", e.ID)
		}
		switch e.Status {
		case "unsupported":
		case "implemented":
			if len(e.Fixtures) == 0 {
				return fmt.Errorf("%s: implemented scope requires conformance fixtures", e.ID)
			}
		default:
			return fmt.Errorf("%s: invalid status %q", e.ID, e.Status)
		}
		if e.Fixtures == nil {
			return fmt.Errorf("%s: fixtures must be an explicit array", e.ID)
		}
		seen := make(map[string]bool, len(e.Fixtures))
		for _, fixture := range e.Fixtures {
			if !fs.ValidPath(fixture) || fixture == "." || strings.Contains(fixture, "\\") || seen[fixture] {
				return fmt.Errorf("%s: invalid or duplicate fixture path %q", e.ID, fixture)
			}
			seen[fixture] = true
			// A prose file or the register itself is not a mechanical fixture.
			if !strings.HasSuffix(fixture, "_test.go") && !strings.Contains("/"+path.Dir(fixture)+"/", "/testdata/") {
				return fmt.Errorf("%s: fixture %q must be a Go test or a testdata file", e.ID, fixture)
			}
			if repo == nil {
				return fmt.Errorf("%s: repository filesystem required to check fixture %q", e.ID, fixture)
			}
			info, err := fs.Stat(repo, fixture)
			if err != nil {
				return fmt.Errorf("%s: fixture %q: %w", e.ID, fixture, err)
			}
			if !info.Mode().IsRegular() || info.Size() == 0 {
				return fmt.Errorf("%s: fixture %q must be a nonempty regular file", e.ID, fixture)
			}
		}
	}
	return nil
}

// VerifyArtifact checks the original PDF's size and SHA-256 without retaining it.
// It never downloads content. Callers provide an opened artifact reader.
func VerifyArtifact(pdf io.Reader) error {
	h := sha256.New()
	n, err := io.Copy(h, pdf)
	if err != nil {
		return fmt.Errorf("read SRD artifact: %w", err)
	}
	if n != artifactBytes || fmt.Sprintf("%x", h.Sum(nil)) != artifactSHA256 {
		return fmt.Errorf("artifact does not match pinned SRD %s size and SHA-256", Version)
	}
	return nil
}
