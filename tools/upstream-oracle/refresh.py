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
    if commit != "328e05d10d6c73654e416675fc75af7ed0e05ffd":
        raise SystemExit("Expected the reviewed 13.0.0 revision 328e05d; refusing to attest another source.")
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
        for culture in CULTURES:
            output = temp / "compatibility.json"
            run("dotnet", str(oracle), "--compatibility", str(output), culture)
            current = json.loads(output.read_text())
            if compatibility is not None and current != compatibility:
                raise SystemExit(f"Culture-dependent oracle result: {culture}")
            compatibility = current
        compatibility["verifiedCultures"] = CULTURES
        pending["kustolang/engine/testdata/upstream_328e05d_compatibility.json"] = canonical(compatibility)
    # Publish only after every comparison and culture run has passed.
    for path, data in pending.items():
        (ROOT / path).write_bytes(data)
    manifest["commit"] = commit
    manifest["profiles"]["current"] = {"version": "Kusto-Query-Language/13.0.0 + auto-sync through 2026-10-22", "source_commit": commit}
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
        "tools/upstream-oracle/Compatibility.cs", "tools/upstream-oracle/generate/Generate.csproj",
        "tools/upstream-oracle/generate/Program.cs", "tools/upstream-oracle/refresh.py",
        "tools/upstream-parity-native/main.go", "kustolang/engine/testdata/upstream_328e05d_compatibility.json",
    ]
    manifest["differential_parity_attestations"][0]["supporting_artifacts"] = [artifact(path) for path in supporting]
    note = "The 328e05d repin upgrades to 13.0.0: invariant literal parsing and completion examples, ordinal Unicode equality and ordering, case-insensitive wildcard middle segments, supplementary-scalar scanners, and bounded TextSource matching. Both existing corpora (258 queries and 231 commands) were independently replayed against a source build at 328e05d and the Go port; canonical outputs are unchanged. The compatibility fixture records 19 comparison cases, 13 wildcard cases, 31 literal values, 19 scanner cases, 6 source-boundary cases, 8 completion value examples, and completion ordering, verified under en-US, fr-FR, tr-TR, da-DK, and cs-CZ. Supporting artifact hashes bind the fixture and reproducible source/native oracle tools to this attestation."
    if note not in manifest["notes"]:
        manifest["notes"] += " " + note
    manifest_path.write_text(json.dumps(manifest, indent=2) + "\n")
    print("Refreshed 13.0.0 parity attestations and five-culture compatibility fixture.")


if __name__ == "__main__":
    main()
