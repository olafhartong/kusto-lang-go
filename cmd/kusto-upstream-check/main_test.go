package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReportsUpstreamCommitRelationships(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		ahead      int
		behind     int
		wantExit   int
		wantOutput string
	}{
		{"identical", "identical", 0, 0, 0, "Microsoft Kusto Query Language is current at head456\n"},
		{"ahead", "ahead", 3, 0, 0, "upstream ahead: pinned pin123, latest head456 (3 ahead)\naffected tracked surfaces: none\nupstream advancement is informational: no changes within configured source paths\n"},
		{"behind", "behind", 0, 2, 1, "upstream behind pinned commit: pinned pin123, latest head456 (2 behind)\naffected tracked surfaces: none\n"},
		{"diverged", "diverged", 4, 2, 1, "upstream diverged from pinned commit: pinned pin123, latest head456 (4 ahead, 2 behind)\naffected tracked surfaces: none\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				assertGitHubRequestHeaders(t, r)
				w.Header().Set("Content-Type", "application/json")
				switch requests {
				case 1:
					if r.URL.Path != "/repos/microsoft/Kusto-Query-Language/commits/HEAD" {
						t.Fatalf("HEAD path = %q", r.URL.Path)
					}
					_, _ = w.Write([]byte(`{"sha":"head456"}`))
				case 2:
					if r.URL.Path != "/repos/microsoft/Kusto-Query-Language/compare/pin123...head456" || r.URL.Query().Get("per_page") != "100" || r.URL.Query().Get("page") != "1" {
						t.Fatalf("compare URL = %q?%s", r.URL.Path, r.URL.RawQuery)
					}
					_, _ = fmt.Fprintf(w, `{"status":%q,"ahead_by":%d,"behind_by":%d,"files":[]}`, test.status, test.ahead, test.behind)
				default:
					t.Fatalf("unexpected request %d", requests)
				}
			}))
			defer server.Close()

			path := writeUpstreamMetadata(t, "pin123")
			var stdout, stderr bytes.Buffer
			if code := run([]string{"-upstream-file", path}, &stdout, &stderr, server.Client(), server.URL); code != test.wantExit {
				t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
			}
			if stdout.String() != test.wantOutput || stderr.Len() != 0 || requests != 2 {
				t.Fatalf("stdout = %q, stderr = %q, requests = %d", stdout.String(), stderr.String(), requests)
			}
		})
	}
}

func TestRunChecksSourcePathsBeforeAcceptingUpstreamAdvancement(t *testing.T) {
	cappedFiles := make([]githubChangedFile, 300)
	for i := range cappedFiles {
		cappedFiles[i].Filename = fmt.Sprintf("docs/file%d.md", i)
	}
	tests := []struct {
		name     string
		files    []githubChangedFile
		paths    []string
		wantExit int
		wantText string
	}{
		{"workflow only", []githubChangedFile{{Filename: ".github/dependabot.yml"}, {Filename: ".github/workflows/codeql.yml"}}, []string{"src/Kusto.Language"}, 0, "advancement is informational"},
		{"unclassified source", []githubChangedFile{{Filename: "src/Kusto.Language/Utils/Thing.cs"}}, []string{"src/Kusto.Language"}, 1, "tracked source changed:"},
		{"rename out of source", []githubChangedFile{{Filename: "docs/Thing.cs", PreviousFilename: "src/Kusto.Language/Utils/Thing.cs"}}, []string{"src/Kusto.Language"}, 1, "tracked source changed:"},
		{"rename into source", []githubChangedFile{{Filename: "src/Kusto.Language/Utils/Thing.cs", PreviousFilename: "docs/Thing.cs"}}, []string{"src/Kusto.Language"}, 1, "tracked source changed:"},
		{"source prefix sibling", []githubChangedFile{{Filename: "src/Kusto.Language.Other/Thing.cs"}}, []string{"src/Kusto.Language"}, 0, "advancement is informational"},
		{"custom source path", []githubChangedFile{{Filename: "custom/Thing.cs"}}, []string{"custom"}, 1, "tracked source changed:"},
		{"file cap", cappedFiles, []string{"src/Kusto.Language"}, 1, "capped at 300 files"},
		{"missing source paths", []githubChangedFile{{Filename: "README.md"}}, nil, 1, "source_paths must not be empty"},
		{"invalid source path", []githubChangedFile{{Filename: "README.md"}}, []string{"../src"}, 1, "source path"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/commits/HEAD") {
					_, _ = w.Write([]byte(`{"sha":"head456"}`))
					return
				}
				ahead, behind := 2, 0
				_ = json.NewEncoder(w).Encode(githubComparison{Status: "ahead", AheadBy: &ahead, BehindBy: &behind, Files: &test.files})
			}))
			defer server.Close()
			path := writeUpstreamMetadata(t, "pin123")
			metadata, err := loadUpstreamMetadata(path)
			if err != nil {
				t.Fatal(err)
			}
			metadata.SourcePaths = test.paths
			data, err := json.Marshal(metadata)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0o600); err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			code := run([]string{"-upstream-file", path}, &stdout, &stderr, server.Client(), server.URL)
			if code != test.wantExit || !strings.Contains(stdout.String()+stderr.String(), test.wantText) {
				t.Fatalf("exit = %d, stdout = %q, stderr = %q", code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestRunReportsMissingPinnedCommit(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if requests == 1 {
			_, _ = w.Write([]byte(`{"sha":"head456"}`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	path := writeUpstreamMetadata(t, "missing123")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-upstream-file", path}, &stdout, &stderr, server.Client(), server.URL); code != 1 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	want := "pinned upstream commit is missing or inaccessible: missing123 (latest head456)\n"
	if stdout.String() != want || stderr.Len() != 0 || requests != 2 {
		t.Fatalf("stdout = %q, stderr = %q, requests = %d", stdout.String(), stderr.String(), requests)
	}
}

func TestRunRejectsMalformedComparisonResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"unknown status", `{"status":"sideways","ahead_by":0,"behind_by":0}`},
		{"missing count", `{"status":"ahead","ahead_by":1}`},
		{"negative count", `{"status":"behind","ahead_by":0,"behind_by":-1}`},
		{"missing files", `{"status":"ahead","ahead_by":1,"behind_by":0}`},
		{"null files", `{"status":"ahead","ahead_by":1,"behind_by":0,"files":null}`},
		{"malformed files", `{"status":"ahead","ahead_by":1,"behind_by":0,"files":{}}`},
		{"empty filename", `{"status":"ahead","ahead_by":1,"behind_by":0,"files":[{"filename":""}]}`},
		{"invalid JSON", `{`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				w.Header().Set("Content-Type", "application/json")
				if requests == 1 {
					_, _ = w.Write([]byte(`{"sha":"head456"}`))
					return
				}
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()

			path := writeUpstreamMetadata(t, "pin123")
			var stdout, stderr bytes.Buffer
			if code := run([]string{"-upstream-file", path}, &stdout, &stderr, server.Client(), server.URL); code != 1 {
				t.Fatalf("exit = %d", code)
			}
			if stdout.Len() != 0 || !strings.Contains(stderr.String(), "check upstream:") || requests != 2 {
				t.Fatalf("stdout = %q, stderr = %q, requests = %d", stdout.String(), stderr.String(), requests)
			}
		})
	}
}

func TestRunClassifiesChangedTrackedSurfaces(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		if requests == 1 {
			_, _ = w.Write([]byte(`{"sha":"head456"}`))
			return
		}
		_, _ = w.Write([]byte(`{
			"status":"ahead",
			"ahead_by":2,
			"behind_by":0,
			"files":[
				{"filename":"src/Kusto.Language/Parser/QueryGrammar.cs"},
				{"filename":"src/Kusto.Language/Functions.cs"},
				{"filename":"src/Kusto.Language/Parser/CommandGrammar.cs"},
				{"filename":"src/Kusto.Language/Diagnostics/DiagnosticFacts.cs"},
				{"filename":"src/Kusto.Language/Legacy/QuickInfoBuilder.cs","previous_filename":"src/Kusto.Language/Editor/QuickInfoBuilder.cs"},
				{"filename":"docs/function-changes.md"}
			]
		}`))
	}))
	defer server.Close()

	path := writeUpstreamMetadata(t, "pin123")
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-upstream-file", path}, &stdout, &stderr, server.Client(), server.URL); code != 1 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	wantSurfaces := "affected tracked surfaces: query grammar/parser/syntax, functions/operators/plugins/aggregates, commands, diagnostics, editor\n"
	if !strings.Contains(stdout.String(), wantSurfaces) || stderr.Len() != 0 || requests != 2 {
		t.Fatalf("stdout = %q, stderr = %q, requests = %d", stdout.String(), stderr.String(), requests)
	}
}

func TestReportAffectedSurfacesWarnsAtGitHubFileCap(t *testing.T) {
	files := make([]githubChangedFile, 300)
	for i := range files {
		files[i].Filename = fmt.Sprintf("src/Other/File%d.cs", i)
	}
	var output bytes.Buffer
	reportAffectedSurfaces(&output, githubComparison{Files: &files}, []string{"src/Kusto.Language"})
	if !strings.Contains(output.String(), "capped at 300 files") {
		t.Fatalf("output = %q", output.String())
	}
}

func TestClassifyChangedFilesIgnoresUntrackedPaths(t *testing.T) {
	files := []githubChangedFile{{Filename: "docs/functions-and-operators.md"}}
	if affected := classifyChangedFiles(files, []string{"src/Kusto.Language", "grammar"}); len(affected) != 0 {
		t.Fatalf("affected = %v", affected)
	}
}

func assertGitHubRequestHeaders(t *testing.T, request *http.Request) {
	t.Helper()
	if request.Header.Get("Accept") != "application/vnd.github+json" ||
		request.Header.Get("X-GitHub-Api-Version") != "2022-11-28" ||
		request.Header.Get("User-Agent") != "kqlanalyzer-upstream-check" {
		t.Fatalf("headers = %#v", request.Header)
	}
}

func TestGitHubRepositoryPartsRejectsUnsupportedRepository(t *testing.T) {
	if _, _, err := githubRepositoryParts("https://example.com/microsoft/Kusto-Query-Language"); err == nil {
		t.Fatal("expected unsupported repository error")
	}
}

func TestVerifyRemoteSourceFiles(t *testing.T) {
	source := []byte("public static class DiagnosticFacts {}")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertGitHubRequestHeaders(t, r)
		if r.URL.Path != "/repos/microsoft/Kusto-Query-Language/contents/src/Kusto.Language/Diagnostics/DiagnosticFacts.cs" || r.URL.Query().Get("ref") != strings.Repeat("a", 40) {
			t.Fatalf("contents URL = %q?%s", r.URL.Path, r.URL.RawQuery)
		}
		_ = json.NewEncoder(w).Encode(githubContent{Type: "file", Encoding: "base64", Content: base64.StdEncoding.EncodeToString(source)})
	}))
	defer server.Close()
	metadata := upstreamMetadata{
		Repository: "https://github.com/microsoft/Kusto-Query-Language",
		Commit:     strings.Repeat("a", 40),
		SourceFiles: map[string]upstreamSourceFile{
			"src/Kusto.Language/Diagnostics/DiagnosticFacts.cs": {SHA256: fmt.Sprintf("%x", sha256.Sum256(source))},
		},
	}
	if err := verifyRemoteSourceFiles(context.Background(), server.Client(), server.URL, metadata, ""); err != nil {
		t.Fatal(err)
	}
	metadata.SourceFiles["src/Kusto.Language/Diagnostics/DiagnosticFacts.cs"] = upstreamSourceFile{SHA256: strings.Repeat("0", 64)}
	if err := verifyRemoteSourceFiles(context.Background(), server.Client(), server.URL, metadata, ""); err == nil || !strings.Contains(err.Error(), "differs from commit") {
		t.Fatalf("error = %v", err)
	}
}

func TestRunVerifiesLocalProvenanceWithoutNetwork(t *testing.T) {
	root, path := writeProvenanceTree(t)
	var stdout, stderr bytes.Buffer
	if code := run([]string{"-upstream-file", path, "-root", root, "-verify-provenance", "-local-only"}, &stdout, &stderr, nil, ""); code != 0 {
		t.Fatalf("exit = %d, stderr = %q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "local Kusto provenance is valid") {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestVerifyLocalProvenanceRejectsTamperedLicense(t *testing.T) {
	root, path := writeProvenanceTree(t)
	metadata, err := loadUpstreamMetadata(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "third_party", "kusto", "LICENSE"), []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyLocalProvenance(root, metadata); err == nil || !strings.Contains(err.Error(), "license hash mismatch") {
		t.Fatalf("error = %v", err)
	}
}

func TestVerifyLocalProvenanceRejectsTamperedNotice(t *testing.T) {
	root, path := writeProvenanceTree(t)
	metadata, err := loadUpstreamMetadata(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "kustolang", "engine", "doc.go"), []byte("https://github.com/microsoft/Kusto-Query-Language Apache-2.0 tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyLocalProvenance(root, metadata); err == nil || !strings.Contains(err.Error(), "notice hash mismatch") {
		t.Fatalf("error = %v", err)
	}
}

func TestVerifyLocalProvenanceRejectsTamperedSourceSnapshot(t *testing.T) {
	root, path := writeProvenanceTree(t)
	metadata, err := loadUpstreamMetadata(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "third_party", "kusto", "source", "DiagnosticFacts.cs"), []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyLocalProvenance(root, metadata); err == nil || !strings.Contains(err.Error(), "source file") || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("error = %v", err)
	}
}

func TestVerifyLocalProvenanceRejectsUnsafeManifestPath(t *testing.T) {
	root, path := writeProvenanceTree(t)
	metadata, err := loadUpstreamMetadata(path)
	if err != nil {
		t.Fatal(err)
	}
	metadata.NoticeFile = "../NOTICE"
	if err := verifyLocalProvenance(root, metadata); err == nil || !strings.Contains(err.Error(), "escapes the repository root") {
		t.Fatalf("error = %v", err)
	}
}

func TestVerifyLocalProvenanceRequiresProfileRevisions(t *testing.T) {
	root, path := writeProvenanceTree(t)
	metadata, err := loadUpstreamMetadata(path)
	if err != nil {
		t.Fatal(err)
	}
	metadata.Profiles["current"] = upstreamProfile{Version: "current", SourceCommit: strings.Repeat("b", 40)}
	if err := verifyLocalProvenance(root, metadata); err == nil || !strings.Contains(err.Error(), "must equal commit") {
		t.Fatalf("error = %v", err)
	}
}

func TestVerifyLocalProvenanceRejectsStaleParityAttestation(t *testing.T) {
	root, path := writeProvenanceTree(t)
	metadata, err := loadUpstreamMetadata(path)
	if err != nil {
		t.Fatal(err)
	}
	metadata.DifferentialParityAttestations[0].Commit = strings.Repeat("b", 40)
	if err := verifyLocalProvenance(root, metadata); err == nil || !strings.Contains(err.Error(), "commit must equal metadata commit") {
		t.Fatalf("error = %v", err)
	}
}

func writeProvenanceTree(t *testing.T) (root, metadataPath string) {
	t.Helper()
	root = t.TempDir()
	licensePath := filepath.Join(root, "third_party", "kusto", "LICENSE")
	noticePath := filepath.Join(root, "kustolang", "engine", "doc.go")
	sourceSnapshotPath := filepath.Join(root, "third_party", "kusto", "source", "DiagnosticFacts.cs")
	corpusPath := filepath.Join(root, "parity-artifacts", "query-corpus.json")
	oracleResultPath := filepath.Join(root, "parity-artifacts", "query-oracle.json")
	nativeResultPath := filepath.Join(root, "parity-artifacts", "query-native.json")
	generatorPath := filepath.Join(root, "tools", "parity-generator.go")
	for _, directory := range []string{
		filepath.Dir(licensePath), filepath.Dir(noticePath), filepath.Dir(sourceSnapshotPath),
		filepath.Dir(corpusPath), filepath.Dir(generatorPath),
	} {
		if err := os.MkdirAll(directory, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	license := []byte("Apache License Version 2.0")
	notice := []byte("https://github.com/microsoft/Kusto-Query-Language Apache-2.0")
	sourceSnapshot := []byte("public static class DiagnosticFacts {}")
	corpus := []byte("[\"print 1\"]\n")
	generator := []byte("package main\n")
	result := []byte("matching parity results")
	for path, data := range map[string][]byte{
		licensePath: license, noticePath: notice, sourceSnapshotPath: sourceSnapshot,
		corpusPath: corpus, generatorPath: generator,
		oracleResultPath: result, nativeResultPath: result,
	} {
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	commit := strings.Repeat("a", 40)
	resultHash := fmt.Sprintf("%x", sha256.Sum256(result))
	metadata := upstreamMetadata{
		Name: "Microsoft Kusto Query Language", Repository: "https://github.com/microsoft/Kusto-Query-Language",
		Commit: commit, License: "Apache-2.0", LicenseFile: "third_party/kusto/LICENSE",
		LicenseSHA256: fmt.Sprintf("%x", sha256.Sum256(license)), NoticeFile: "kustolang/engine/doc.go",
		NoticeSHA256: fmt.Sprintf("%x", sha256.Sum256(notice)), SourcePaths: []string{"src/Kusto.Language", "grammar"},
		SourceFiles: map[string]upstreamSourceFile{
			"src/Kusto.Language/Diagnostics/DiagnosticFacts.cs": {
				SnapshotFile: "third_party/kusto/source/DiagnosticFacts.cs",
				SHA256:       fmt.Sprintf("%x", sha256.Sum256(sourceSnapshot)),
			},
		},
		Profiles: map[string]upstreamProfile{"current": {Version: "current", SourceCommit: commit}},
		DifferentialParityAttestations: []differentialParityAttestation{{
			Name: "query-structural", Commit: commit, QueryCount: 1,
			Corpus:       parityArtifact{Path: "parity-artifacts/query-corpus.json", SHA256: fmt.Sprintf("%x", sha256.Sum256(corpus))},
			Generator:    parityArtifact{Path: "tools/parity-generator.go", SHA256: fmt.Sprintf("%x", sha256.Sum256(generator))},
			OracleResult: parityArtifact{Path: "parity-artifacts/query-oracle.json", SHA256: resultHash},
			NativeResult: parityArtifact{Path: "parity-artifacts/query-native.json", SHA256: resultHash},
		}},
	}
	data, err := json.Marshal(metadata)
	if err != nil {
		t.Fatal(err)
	}
	metadataPath = filepath.Join(root, "third_party", "kusto", "UPSTREAM.json")
	if err := os.WriteFile(metadataPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return root, metadataPath
}

func TestVerifyLocalProvenanceRejectsParityArtifactHashMismatch(t *testing.T) {
	for _, artifactName := range []string{"corpus", "generator"} {
		t.Run(artifactName, func(t *testing.T) {
			root, path := writeProvenanceTree(t)
			metadata, err := loadUpstreamMetadata(path)
			if err != nil {
				t.Fatal(err)
			}
			artifact := metadata.DifferentialParityAttestations[0].Corpus
			if artifactName == "generator" {
				artifact = metadata.DifferentialParityAttestations[0].Generator
			}
			artifactPath := filepath.Join(root, filepath.FromSlash(artifact.Path))
			if err := os.WriteFile(artifactPath, []byte("tampered"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := verifyLocalProvenance(root, metadata); err == nil || !strings.Contains(err.Error(), artifactName+" hash mismatch") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestVerifyLocalProvenanceRejectsMalformedParityAttestations(t *testing.T) {
	tests := []struct {
		name string
		edit func(*upstreamMetadata)
		want string
	}{
		{"missing", func(metadata *upstreamMetadata) { metadata.DifferentialParityAttestations = nil }, "must not be empty"},
		{"zero count", func(metadata *upstreamMetadata) { metadata.DifferentialParityAttestations[0].QueryCount = 0 }, "query_count must be positive"},
		{"wrong count", func(metadata *upstreamMetadata) { metadata.DifferentialParityAttestations[0].QueryCount = 2 }, "corpus contains 1 queries"},
		{"unsafe path", func(metadata *upstreamMetadata) {
			metadata.DifferentialParityAttestations[0].Generator.Path = "../generator.go"
		}, "escapes the repository root"},
		{"bad result hash", func(metadata *upstreamMetadata) {
			metadata.DifferentialParityAttestations[0].OracleResult.SHA256 = "bad"
		}, "oracle result sha256"},
		{"parity mismatch", func(metadata *upstreamMetadata) {
			metadata.DifferentialParityAttestations[0].NativeResult.SHA256 = strings.Repeat("f", 64)
		}, "native result hash mismatch"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, path := writeProvenanceTree(t)
			metadata, err := loadUpstreamMetadata(path)
			if err != nil {
				t.Fatal(err)
			}
			test.edit(&metadata)
			if err := verifyLocalProvenance(root, metadata); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func writeUpstreamMetadata(t *testing.T, commit string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "UPSTREAM.json")
	data := []byte(`{"name":"Microsoft Kusto Query Language","repository":"https://github.com/microsoft/Kusto-Query-Language","commit":"` + commit + `","source_paths":["src/Kusto.Language","src/Kusto.Language.Generators","grammar"]}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
