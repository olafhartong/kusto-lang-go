package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const githubAPIBase = "https://api.github.com"

type upstreamMetadata struct {
	Name                           string                          `json:"name"`
	Repository                     string                          `json:"repository"`
	Commit                         string                          `json:"commit"`
	License                        string                          `json:"license"`
	LicenseFile                    string                          `json:"license_file"`
	LicenseSHA256                  string                          `json:"license_sha256"`
	NoticeFile                     string                          `json:"notice_file"`
	NoticeSHA256                   string                          `json:"notice_sha256"`
	SourcePaths                    []string                        `json:"source_paths"`
	SourceFiles                    map[string]upstreamSourceFile   `json:"source_files"`
	Profiles                       map[string]upstreamProfile      `json:"profiles"`
	DifferentialParityAttestations []differentialParityAttestation `json:"differential_parity_attestations"`
}

type upstreamSourceFile struct {
	SnapshotFile string `json:"snapshot_file"`
	SHA256       string `json:"sha256"`
}

type upstreamProfile struct {
	Version      string `json:"version"`
	SourceCommit string `json:"source_commit"`
}

type differentialParityAttestation struct {
	Name                string           `json:"name"`
	Commit              string           `json:"commit"`
	QueryCount          int              `json:"query_count"`
	Corpus              parityArtifact   `json:"corpus"`
	Generator           parityArtifact   `json:"generator"`
	OracleResult        parityArtifact   `json:"oracle_result"`
	NativeResult        parityArtifact   `json:"native_result"`
	SupportingArtifacts []parityArtifact `json:"supporting_artifacts,omitempty"`
}

type parityArtifact struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type githubCommit struct {
	SHA string `json:"sha"`
}

type githubComparison struct {
	Status   string               `json:"status"`
	AheadBy  *int                 `json:"ahead_by"`
	BehindBy *int                 `json:"behind_by"`
	Files    *[]githubChangedFile `json:"files"`
}

type githubChangedFile struct {
	Filename         string `json:"filename"`
	PreviousFilename string `json:"previous_filename"`
}

type githubContent struct {
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, http.DefaultClient, githubAPIBase))
}

func run(args []string, stdout, stderr io.Writer, client *http.Client, apiBase string) int {
	flags := flag.NewFlagSet("kusto-upstream-check", flag.ContinueOnError)
	flags.SetOutput(stderr)
	upstreamFile := flags.String("upstream-file", "third_party/kusto/UPSTREAM.json", "Path to the pinned upstream metadata file.")
	root := flags.String("root", ".", "Repository root used for local provenance paths.")
	verifyProvenance := flags.Bool("verify-provenance", false, "Validate local license, notice, profile, and source-path provenance.")
	localOnly := flags.Bool("local-only", false, "Validate local provenance without querying GitHub.")
	verifySourceFiles := flags.Bool("verify-source-files", false, "Compare pinned source snapshots with files at the recorded upstream commit.")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	metadata, err := loadUpstreamMetadata(*upstreamFile)
	if err != nil {
		fmt.Fprintf(stderr, "load upstream metadata: %v\n", err)
		return 1
	}
	if *verifyProvenance || *localOnly {
		if err := verifyLocalProvenance(*root, metadata); err != nil {
			fmt.Fprintf(stderr, "verify local provenance: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "local Kusto provenance is valid")
	}
	if *localOnly {
		return 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if *verifySourceFiles {
		if err := verifyRemoteSourceFiles(ctx, client, apiBase, metadata, os.Getenv("GITHUB_TOKEN")); err != nil {
			fmt.Fprintf(stderr, "verify upstream source files: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "pinned Kusto source snapshots match upstream")
	}
	latest, err := latestGitHubCommit(ctx, client, apiBase, metadata.Repository, os.Getenv("GITHUB_TOKEN"))
	if err != nil {
		fmt.Fprintf(stderr, "check upstream: %v\n", err)
		return 1
	}

	comparison, missingPin, err := compareGitHubCommits(ctx, client, apiBase, metadata.Repository, metadata.Commit, latest, os.Getenv("GITHUB_TOKEN"))
	if err != nil {
		fmt.Fprintf(stderr, "check upstream: %v\n", err)
		return 1
	}
	if missingPin {
		fmt.Fprintf(stdout, "pinned upstream commit is missing or inaccessible: %s (latest %s)\n", metadata.Commit, latest)
		return 1
	}

	name := metadata.Name
	if name == "" {
		name = metadata.Repository
	}
	switch comparison.Status {
	case "identical":
		fmt.Fprintf(stdout, "%s is current at %s\n", name, latest)
		return 0
	case "ahead":
		fmt.Fprintf(stdout, "upstream ahead: pinned %s, latest %s (%d ahead)\n", metadata.Commit, latest, *comparison.AheadBy)
	case "behind":
		fmt.Fprintf(stdout, "upstream behind pinned commit: pinned %s, latest %s (%d behind)\n", metadata.Commit, latest, *comparison.BehindBy)
	case "diverged":
		fmt.Fprintf(stdout, "upstream diverged from pinned commit: pinned %s, latest %s (%d ahead, %d behind)\n", metadata.Commit, latest, *comparison.AheadBy, *comparison.BehindBy)
	}
	reportAffectedSurfaces(stdout, comparison, metadata.SourcePaths)
	if comparison.Status != "ahead" || len(*comparison.Files) >= 300 {
		return 1
	}
	// A complete file list and valid source scope are required before treating
	// upstream advancement as informational. Surface categories are only labels.
	if len(metadata.SourcePaths) == 0 {
		fmt.Fprintln(stderr, "check upstream: source_paths must not be empty")
		return 1
	}
	for _, path := range metadata.SourcePaths {
		if err := validateRelativePath(path); err != nil {
			fmt.Fprintf(stderr, "check upstream: source path %q: %v\n", path, err)
			return 1
		}
	}
	for _, file := range *comparison.Files {
		if isTrackedSourcePath(file.Filename, metadata.SourcePaths) || isTrackedSourcePath(file.PreviousFilename, metadata.SourcePaths) {
			fmt.Fprintf(stdout, "tracked source changed: %s\n", file.Filename)
			return 1
		}
	}
	fmt.Fprintln(stdout, "upstream advancement is informational: no changes within configured source paths")
	return 0
}

func verifyRemoteSourceFiles(ctx context.Context, client *http.Client, apiBase string, metadata upstreamMetadata, token string) error {
	if len(metadata.SourceFiles) == 0 {
		return fmt.Errorf("source_files must not be empty")
	}
	owner, name, err := githubRepositoryParts(metadata.Repository)
	if err != nil {
		return err
	}
	for sourcePath, sourceFile := range metadata.SourceFiles {
		segments := strings.Split(sourcePath, "/")
		for i := range segments {
			segments[i] = url.PathEscape(segments[i])
		}
		endpoint := strings.TrimRight(apiBase, "/") + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name) + "/contents/" + strings.Join(segments, "/") + "?ref=" + url.QueryEscape(metadata.Commit)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return err
		}
		setGitHubHeaders(req, token)
		response, err := client.Do(req)
		if err != nil {
			return err
		}
		body, readErr := io.ReadAll(io.LimitReader(response.Body, 2<<20))
		closeErr := response.Body.Close()
		if readErr != nil {
			return readErr
		}
		if closeErr != nil {
			return closeErr
		}
		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("GitHub contents API returned %s for %s: %s", response.Status, sourcePath, strings.TrimSpace(string(body)))
		}
		var content githubContent
		if err := json.Unmarshal(body, &content); err != nil {
			return err
		}
		if content.Type != "file" || content.Encoding != "base64" {
			return fmt.Errorf("GitHub contents API returned unsupported content for %s", sourcePath)
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(content.Content, "\n", ""))
		if err != nil {
			return fmt.Errorf("decode %s: %w", sourcePath, err)
		}
		actualHash := fmt.Sprintf("%x", sha256.Sum256(decoded))
		if actualHash != sourceFile.SHA256 {
			return fmt.Errorf("source file %q differs from commit %s: manifest %s, upstream %s", sourcePath, metadata.Commit, sourceFile.SHA256, actualHash)
		}
	}
	return nil
}

var (
	commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

func verifyLocalProvenance(root string, metadata upstreamMetadata) error {
	if metadata.License != "Apache-2.0" {
		return fmt.Errorf("license must be Apache-2.0")
	}
	if !commitPattern.MatchString(metadata.Commit) {
		return fmt.Errorf("commit must be a lowercase 40-character SHA")
	}
	current, ok := metadata.Profiles["current"]
	if !ok || current.SourceCommit != metadata.Commit {
		return fmt.Errorf("profiles.current.source_commit must equal commit")
	}
	if len(metadata.SourcePaths) == 0 {
		return fmt.Errorf("source_paths must not be empty")
	}
	for _, path := range metadata.SourcePaths {
		if err := validateRelativePath(path); err != nil {
			return fmt.Errorf("source path %q: %w", path, err)
		}
	}
	if err := verifyDifferentialParityAttestations(root, metadata); err != nil {
		return err
	}
	if len(metadata.SourceFiles) == 0 {
		return fmt.Errorf("source_files must not be empty")
	}
	for sourcePath, sourceFile := range metadata.SourceFiles {
		if err := validateRelativePath(sourcePath); err != nil {
			return fmt.Errorf("source file %q: %w", sourcePath, err)
		}
		if sourceFile.SnapshotFile == "" || sourceFile.SHA256 == "" {
			return fmt.Errorf("source file %q must record snapshot_file and sha256", sourcePath)
		}
		snapshotPath, err := localProvenancePath(root, sourceFile.SnapshotFile)
		if err != nil {
			return fmt.Errorf("source file %q snapshot_file: %w", sourcePath, err)
		}
		data, err := os.ReadFile(snapshotPath)
		if err != nil {
			return fmt.Errorf("read source file %q snapshot: %w", sourcePath, err)
		}
		actualHash := fmt.Sprintf("%x", sha256.Sum256(data))
		if actualHash != sourceFile.SHA256 {
			return fmt.Errorf("source file %q hash mismatch: manifest %s, actual %s", sourcePath, sourceFile.SHA256, actualHash)
		}
	}

	licensePath, err := localProvenancePath(root, metadata.LicenseFile)
	if err != nil {
		return fmt.Errorf("license_file: %w", err)
	}
	licenseData, err := os.ReadFile(licensePath)
	if err != nil {
		return fmt.Errorf("read license_file: %w", err)
	}
	if metadata.LicenseSHA256 == "" {
		return fmt.Errorf("license_sha256 is required")
	}
	actualLicenseHash := fmt.Sprintf("%x", sha256.Sum256(licenseData))
	if actualLicenseHash != metadata.LicenseSHA256 {
		return fmt.Errorf("license hash mismatch: manifest %s, actual %s", metadata.LicenseSHA256, actualLicenseHash)
	}

	noticePath, err := localProvenancePath(root, metadata.NoticeFile)
	if err != nil {
		return fmt.Errorf("notice_file: %w", err)
	}
	noticeData, err := os.ReadFile(noticePath)
	if err != nil {
		return fmt.Errorf("read notice_file: %w", err)
	}
	notice := string(noticeData)
	if !strings.Contains(notice, metadata.Repository) || !strings.Contains(notice, metadata.License) {
		return fmt.Errorf("notice_file must name the upstream repository and %s license", metadata.License)
	}
	if metadata.NoticeSHA256 == "" {
		return fmt.Errorf("notice_sha256 is required")
	}
	actualNoticeHash := fmt.Sprintf("%x", sha256.Sum256(noticeData))
	if actualNoticeHash != metadata.NoticeSHA256 {
		return fmt.Errorf("notice hash mismatch: manifest %s, actual %s", metadata.NoticeSHA256, actualNoticeHash)
	}
	return verifyUpstreamChangelog(root, metadata)
}

func verifyUpstreamChangelog(root string, metadata upstreamMetadata) error {
	const name = "UPSTREAM_CHANGELOG.md"
	path, err := localProvenancePath(root, name)
	if err != nil {
		return fmt.Errorf("upstream changelog path: %w", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read upstream changelog: %w", err)
	}
	link := strings.TrimRight(metadata.Repository, "/") + "/commit/" + metadata.Commit
	const header = "| Area | Impacted functions or symbols | Confirmed behavior or update |"
	// Scope the table to the pinned revision's entry. A revision mentioned in
	// the introduction or an unrelated historical table is not enough.
	for _, section := range strings.Split("\n"+string(data), "\n## ")[1:] {
		if !strings.Contains(section, "]("+link+")") {
			continue
		}
		_, table, found := strings.Cut(section, header)
		if found {
			for _, line := range strings.Split(table, "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.Trim(line, "| :-\t") == "" {
					continue
				}
				if !strings.HasPrefix(line, "|") {
					break
				}
				cells := strings.SplitN(line, "|", 5)
				if len(cells) == 5 && strings.TrimSpace(cells[1]) != "" && strings.TrimSpace(cells[2]) != "" && strings.TrimSpace(cells[3]) != "" {
					return nil
				}
			}
		}
		return fmt.Errorf("%s entry for %s must contain a populated Area / Impacted functions or symbols / Confirmed behavior or update table", name, metadata.Commit)
	}
	return fmt.Errorf("%s must include an entry linking to %s; refresh the changelog with each upstream compatibility update", name, link)
}

func verifyDifferentialParityAttestations(root string, metadata upstreamMetadata) error {
	if len(metadata.DifferentialParityAttestations) == 0 {
		return fmt.Errorf("differential_parity_attestations must not be empty")
	}
	seen := make(map[string]struct{}, len(metadata.DifferentialParityAttestations))
	for i, attestation := range metadata.DifferentialParityAttestations {
		label := fmt.Sprintf("differential parity attestation %d", i)
		if attestation.Name == "" {
			return fmt.Errorf("%s name is required", label)
		}
		if _, ok := seen[attestation.Name]; ok {
			return fmt.Errorf("differential parity attestation name %q is duplicated", attestation.Name)
		}
		seen[attestation.Name] = struct{}{}
		label = fmt.Sprintf("differential parity attestation %q", attestation.Name)
		if attestation.Commit != metadata.Commit {
			return fmt.Errorf("%s commit must equal metadata commit", label)
		}
		if attestation.QueryCount <= 0 {
			return fmt.Errorf("%s query_count must be positive", label)
		}
		if err := verifyParityArtifact(root, label+" corpus", attestation.Corpus); err != nil {
			return err
		}
		if err := verifyParityCorpusCount(root, label, attestation.Corpus, attestation.QueryCount); err != nil {
			return err
		}
		if err := verifyParityArtifact(root, label+" generator", attestation.Generator); err != nil {
			return err
		}
		if err := verifyParityArtifact(root, label+" oracle result", attestation.OracleResult); err != nil {
			return err
		}
		if err := verifyParityArtifact(root, label+" native result", attestation.NativeResult); err != nil {
			return err
		}
		if attestation.OracleResult.SHA256 != attestation.NativeResult.SHA256 {
			return fmt.Errorf("%s oracle and native result hashes do not match", label)
		}
		for _, artifact := range attestation.SupportingArtifacts {
			if err := verifyParityArtifact(root, label+" supporting artifact", artifact); err != nil {
				return err
			}
		}
	}
	return nil
}

func verifyParityCorpusCount(root, label string, artifact parityArtifact, want int) error {
	path, err := localProvenancePath(root, artifact.Path)
	if err != nil {
		return fmt.Errorf("%s corpus path: %w", label, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s corpus: %w", label, err)
	}
	var queries []string
	if err := json.Unmarshal(data, &queries); err != nil {
		return fmt.Errorf("decode %s corpus: %w", label, err)
	}
	if len(queries) != want {
		return fmt.Errorf("%s query_count is %d, corpus contains %d queries", label, want, len(queries))
	}
	return nil
}

func verifyParityArtifact(root, label string, artifact parityArtifact) error {
	if err := validateRelativePath(artifact.Path); err != nil {
		return fmt.Errorf("%s path: %w", label, err)
	}
	if !sha256Pattern.MatchString(artifact.SHA256) {
		return fmt.Errorf("%s sha256 must be a lowercase SHA-256", label)
	}
	path, err := localProvenancePath(root, artifact.Path)
	if err != nil {
		return fmt.Errorf("%s path: %w", label, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat %s: %w", label, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s path must identify a regular file when present", label)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", label, err)
	}
	actualHash := fmt.Sprintf("%x", sha256.Sum256(data))
	if actualHash != artifact.SHA256 {
		return fmt.Errorf("%s hash mismatch: manifest %s, actual %s", label, artifact.SHA256, actualHash)
	}
	return nil
}

func localProvenancePath(root, path string) (string, error) {
	if err := validateRelativePath(path); err != nil {
		return "", err
	}
	return filepath.Join(root, filepath.FromSlash(path)), nil
}

func validateRelativePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}
	if filepath.IsAbs(path) {
		return fmt.Errorf("absolute paths are not allowed")
	}
	clean := filepath.Clean(filepath.FromSlash(path))
	if clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path escapes the repository root")
	}
	return nil
}

func loadUpstreamMetadata(path string) (upstreamMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return upstreamMetadata{}, err
	}
	var metadata upstreamMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return upstreamMetadata{}, err
	}
	if metadata.Repository == "" {
		return upstreamMetadata{}, fmt.Errorf("repository is required")
	}
	if metadata.Commit == "" {
		return upstreamMetadata{}, fmt.Errorf("commit is required")
	}
	return metadata, nil
}

func latestGitHubCommit(ctx context.Context, client *http.Client, apiBase, repository, token string) (string, error) {
	owner, name, err := githubRepositoryParts(repository)
	if err != nil {
		return "", err
	}
	endpoint := strings.TrimRight(apiBase, "/") + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name) + "/commits/HEAD"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	setGitHubHeaders(req, token)

	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return "", fmt.Errorf("GitHub API returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}

	var commit githubCommit
	if err := json.NewDecoder(response.Body).Decode(&commit); err != nil {
		return "", err
	}
	if commit.SHA == "" {
		return "", fmt.Errorf("GitHub API response did not include a commit SHA")
	}
	return commit.SHA, nil
}

func compareGitHubCommits(ctx context.Context, client *http.Client, apiBase, repository, base, head, token string) (comparison githubComparison, missingBase bool, err error) {
	owner, name, err := githubRepositoryParts(repository)
	if err != nil {
		return githubComparison{}, false, err
	}
	comparisonPath := url.PathEscape(base + "..." + head)
	endpoint := strings.TrimRight(apiBase, "/") + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name) + "/compare/" + comparisonPath + "?per_page=100&page=1"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return githubComparison{}, false, err
	}
	setGitHubHeaders(req, token)

	response, err := client.Do(req)
	if err != nil {
		return githubComparison{}, false, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode == http.StatusNotFound {
		return githubComparison{}, true, nil
	}
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return githubComparison{}, false, fmt.Errorf("GitHub compare API returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	if err := json.NewDecoder(response.Body).Decode(&comparison); err != nil {
		return githubComparison{}, false, err
	}
	if comparison.AheadBy == nil || comparison.BehindBy == nil || *comparison.AheadBy < 0 || *comparison.BehindBy < 0 {
		return githubComparison{}, false, fmt.Errorf("GitHub compare API response has invalid ahead_by or behind_by counts")
	}
	switch comparison.Status {
	case "ahead", "behind", "diverged", "identical":
	default:
		return githubComparison{}, false, fmt.Errorf("GitHub compare API response has unknown status %q", comparison.Status)
	}
	if comparison.Status != "identical" && comparison.Files == nil {
		return githubComparison{}, false, fmt.Errorf("GitHub compare API response did not include changed files")
	}
	if comparison.Files == nil {
		empty := make([]githubChangedFile, 0)
		comparison.Files = &empty
	}
	if len(*comparison.Files) > 300 {
		return githubComparison{}, false, fmt.Errorf("GitHub compare API response exceeded the 300-file limit")
	}
	for _, file := range *comparison.Files {
		if strings.TrimSpace(file.Filename) == "" {
			return githubComparison{}, false, fmt.Errorf("GitHub compare API response included a changed file without a filename")
		}
	}
	return comparison, false, nil
}

var trackedSurfaceClassifiers = []struct {
	name  string
	match func(string) bool
}{
	{
		name: "query grammar/parser/syntax",
		match: func(path string) bool {
			return strings.Contains(path, "/grammar/") || strings.Contains(path, "/parser/") ||
				strings.Contains(path, "/syntax/") || strings.Contains(path, "grammar.cs") ||
				strings.Contains(path, "parser.cs") || strings.Contains(path, "syntaxfacts.cs")
		},
	},
	{
		name: "functions/operators/plugins/aggregates",
		match: func(path string) bool {
			return strings.Contains(path, "function") || strings.Contains(path, "operator") ||
				strings.Contains(path, "plugin") || strings.Contains(path, "aggregate")
		},
	},
	{
		name: "commands",
		match: func(path string) bool {
			return strings.Contains(path, "command")
		},
	},
	{
		name: "diagnostics",
		match: func(path string) bool {
			return strings.Contains(path, "diagnostic")
		},
	},
	{
		name: "editor",
		match: func(path string) bool {
			return strings.Contains(path, "/editor/") || strings.Contains(path, "editorsyntax")
		},
	},
}

func classifyChangedFiles(files []githubChangedFile, sourcePaths []string) []string {
	affected := make([]string, 0, len(trackedSurfaceClassifiers))
	for _, classifier := range trackedSurfaceClassifiers {
		for _, file := range files {
			for _, filename := range []string{file.Filename, file.PreviousFilename} {
				if !isTrackedSourcePath(filename, sourcePaths) {
					continue
				}
				path := "/" + strings.ToLower(strings.ReplaceAll(filename, "\\", "/"))
				if classifier.match(path) {
					affected = append(affected, classifier.name)
					break
				}
			}
			if len(affected) > 0 && affected[len(affected)-1] == classifier.name {
				break
			}
		}
	}
	return affected
}

func isTrackedSourcePath(filename string, sourcePaths []string) bool {
	filename = strings.Trim(strings.ToLower(strings.ReplaceAll(filename, "\\", "/")), "/")
	if filename == "" {
		return false
	}
	for _, sourcePath := range sourcePaths {
		sourcePath = strings.Trim(strings.ToLower(strings.ReplaceAll(sourcePath, "\\", "/")), "/")
		if filename == sourcePath || strings.HasPrefix(filename, sourcePath+"/") {
			return true
		}
	}
	return false
}

func reportAffectedSurfaces(w io.Writer, comparison githubComparison, sourcePaths []string) {
	affected := classifyChangedFiles(*comparison.Files, sourcePaths)
	if len(affected) == 0 {
		fmt.Fprintln(w, "affected tracked surfaces: none")
	} else {
		fmt.Fprintf(w, "affected tracked surfaces: %s\n", strings.Join(affected, ", "))
	}
	if len(*comparison.Files) == 300 {
		fmt.Fprintln(w, "changed-file classification may be incomplete: GitHub compare results are capped at 300 files")
	}
}

func setGitHubHeaders(req *http.Request, token string) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "kqlanalyzer-upstream-check")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func githubRepositoryParts(repository string) (string, string, error) {
	parsed, err := url.Parse(repository)
	if err != nil {
		return "", "", err
	}
	if !strings.EqualFold(parsed.Hostname(), "github.com") {
		return "", "", fmt.Errorf("unsupported upstream repository %q", repository)
	}
	parts := strings.Split(strings.Trim(strings.TrimSuffix(parsed.Path, ".git"), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid GitHub repository %q", repository)
	}
	return parts[0], parts[1], nil
}
