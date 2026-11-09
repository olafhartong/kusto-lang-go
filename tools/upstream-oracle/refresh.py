#!/usr/bin/env python3
"""Rebuild the pinned source oracle and refresh verified parity evidence."""
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile

ROOT = Path(__file__).resolve().parents[2]
CULTURES = ["en-US", "fr-FR", "tr-TR", "da-DK", "cs-CZ"]


def run(*args, **kwargs):
    subprocess.run(args, cwd=ROOT, check=True, **kwargs)


def canonical(value):
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n").encode()


def artifact(path):
    return {"path": path, "sha256": hashlib.sha256((ROOT / path).read_bytes()).hexdigest()}


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: python3 tools/upstream-oracle/refresh.py /path/to/Kusto-Query-Language")
    source = Path(sys.argv[1]).resolve()
    commit = subprocess.check_output(["git", "-C", str(source), "rev-parse", "HEAD"], text=True).strip()
    if commit != "dbd7ca7684adfd7a38536f74728670dc54a67acc":
        raise SystemExit("Expected the reviewed 13.0.0 revision dbd7ca7; refusing to attest another source.")
    run("git", "-C", str(source), "diff", "--exit-code", "HEAD")
    manifest_path = ROOT / "third_party/kusto/UPSTREAM.json"
    manifest = json.loads(manifest_path.read_text())
    # Existing vendored source must remain byte-identical to the reviewed source.
    for path, info in manifest["source_files"].items():
        if (source / path).read_bytes() != (ROOT / info["snapshot_file"]).read_bytes():
            raise SystemExit(f"Source snapshot changed: {path}; update and review it before refreshing.")
    pending = {}
    with tempfile.TemporaryDirectory(prefix="kusto-source-oracle-") as temp:
        temp = Path(temp)
        prop = f"-p:UpstreamRoot={source}"
        generated = temp / "generated"
        run("dotnet", "run", "--project", "tools/upstream-oracle/generate/Generate.csproj", prop, "--", str(generated))
        run("dotnet", "build", "tools/upstream-oracle/Oracle.csproj", prop, f"-p:GeneratedRoot={generated}")
        oracle = ROOT / "tools/upstream-oracle/bin/Debug/net10.0/Oracle.dll"
        native = temp / "native"
        run("go", "build", "-o", str(native), "./tools/upstream-parity-native")
        for attestation in manifest["differential_parity_attestations"]:
            corpus = ROOT / attestation["corpus"]["path"]
            oracle_result, native_result = temp / "oracle.json", temp / "native.json"
            run("dotnet", str(oracle), str(corpus), str(oracle_result))
            run(str(native), str(corpus), str(native_result))
            expected = json.loads(oracle_result.read_text())
            actual = json.loads(native_result.read_text())
            if expected != actual:
                for i, (want, got) in enumerate(zip(expected, actual)):
                    if want != got:
                        raise SystemExit(f"Parity failed for {attestation['name']} query {i}: {want['query']!r}\nsource: {want}\nnative: {got}")
                raise SystemExit(f"Result count mismatch: {attestation['name']}")
            if len(expected) != attestation["query_count"]:
                raise SystemExit(f"Unexpected corpus count: {attestation['name']}")
            for key in ["oracle_result", "native_result"]:
                pending[attestation[key]["path"]] = canonical(expected)
            attestation["commit"] = commit
            print(f"Matched {len(expected)} cases: {attestation['name']}", flush=True)
        compatibility = None
        caching = None
        november = None
        for culture in CULTURES:
            output = temp / "compatibility.json"
            run("dotnet", str(oracle), "--compatibility", str(output), culture, commit)
            current = json.loads(output.read_text())
            if compatibility is not None and current != compatibility:
                raise SystemExit(f"Culture-dependent oracle result: {culture}")
            compatibility = current
            run("dotnet", str(oracle), "--caching", str(output), culture, commit)
            current = json.loads(output.read_text())
            if caching is not None and current != caching:
                raise SystemExit(f"Culture-dependent caching oracle result: {culture}")
            caching = current
            run("dotnet", str(oracle), "--november", str(output), culture, commit)
            current = json.loads(output.read_text())
            if november is not None and current != november:
                raise SystemExit(f"Culture-dependent November oracle result: {culture}")
            november = current
        november["verifiedCultures"] = CULTURES
        behavior_path = f"kustolang/engine/testdata/upstream_{commit[:7]}_behavior.json"
        behavior = ROOT / behavior_path
        previous = behavior.read_bytes() if behavior.exists() else None
        try:
            behavior.write_bytes(canonical(november))
            run("go", "test", "./kustolang/engine", "-run", "TestUpstreamNov08")
        except BaseException:
            if previous is None:
                behavior.unlink(missing_ok=True)
            else:
                behavior.write_bytes(previous)
            raise
        pending[behavior_path] = canonical(november)
        compatibility["verifiedCultures"] = CULTURES
        caching["verifiedCultures"] = CULTURES
        pending[f"kustolang/engine/testdata/upstream_{commit[:7]}_compatibility.json"] = canonical(compatibility)
        pending[f"kustolang/engine/testdata/upstream_{commit[:7]}_caching.json"] = canonical(caching)
    # Publish only after every comparison and culture run has passed.
    for path, data in pending.items():
        (ROOT / path).write_bytes(data)
    manifest["commit"] = commit
    manifest["profiles"]["current"] = {"version": "Kusto-Query-Language/13.0.0 + auto-sync through 2026-11-08", "source_commit": commit}
    notice = ROOT / manifest["notice_file"]
    text = notice.read_text()
    import re
    text = re.sub(r"(Current source revision: `)[0-9a-f]{40}(`)", rf"\g<1>{commit}\2", text)
    notice.write_text(text)
    manifest["notice_sha256"] = hashlib.sha256(notice.read_bytes()).hexdigest()
    for attestation in manifest["differential_parity_attestations"]:
        for key in ["corpus", "generator", "oracle_result", "native_result"]:
            attestation[key] = artifact(attestation[key]["path"])
    supporting = [
        "tools/upstream-oracle/Oracle.csproj", "tools/upstream-oracle/Program.cs",
        "tools/upstream-oracle/Compatibility.cs", "tools/upstream-oracle/Caching.cs", "tools/upstream-oracle/November.cs", "tools/upstream-oracle/generate/Generate.csproj",
        "tools/upstream-oracle/generate/Program.cs", "tools/upstream-oracle/refresh.py",
        "tools/upstream-parity-native/main.go", "kustolang/engine/testdata/upstream_328e05d_compatibility.json",
        "kustolang/engine/testdata/upstream_82267c7_compatibility.json",
        "kustolang/engine/testdata/upstream_82267c7_caching.json",
        f"kustolang/engine/testdata/upstream_{commit[:7]}_behavior.json",
        f"kustolang/engine/testdata/upstream_{commit[:7]}_compatibility.json",
        f"kustolang/engine/testdata/upstream_{commit[:7]}_caching.json",
    ]
    manifest["differential_parity_attestations"][0]["supporting_artifacts"] = [artifact(path) for path in supporting]
    note = "The dbd7ca7 repin covers source changes from 2026-11-02 through 2026-11-08 without a Microsoft package-version bump. Added the DateTimeLiteral analyzer (KS517 invalid literals and KS518 ambiguous numeric literals) and an internal service-literal interpretation parser; SyntaxToken.Value retains upstream's separate invariant token parser. Added eight graph_model snapshot_indexing/graph_query policy commands, command-completion metadata, and generic unknown-name diagnostics for graph-model references. Keyword-as-identifier function names are now recognized by the grammar parser. Complete commands skip partial-recovery search. Code-service binding/diagnostics/analyzers propagate cancellation without caching it; deferred computed-function expansion uses an independent token. ParseAndAnalyze and has_any/has_all AST kinds already matched and require no runtime change. The CodeGenerator unindented-output option and .NET span-based timespan parsing are upstream tooling/allocation improvements; existing Go string-slice and literal behavior is preserved. Both existing differential corpora (489 cases) were replayed from this exact source revision, plus a dedicated November datetime/analyzer/parser/command fixture under five cultures. Historical fixtures and their evidence remain checked in."
    if note not in manifest["notes"]:
        manifest["notes"] += " " + note
    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n")
    print("Refreshed 13.0.0 parity attestations and five-culture compatibility fixture.")


if __name__ == "__main__":
    main()
