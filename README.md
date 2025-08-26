# Monorepo Dependency Diff Tool — Plan & TODOs

> A pragmatic, actionable plan to build an OSS CLI that computes which packages in a monorepo changed between two commits and the *impact matrix* (which other packages are affected and which tests/builds to run). Designed for easy CI integration and incremental adoption.

---

## 1) One-paragraph overview

A fast, developer-friendly CLI and library that analyzes a monorepo to report: which packages changed (by git diff), the transitive dependency impact (which packages depend on changed ones), and a prioritized list of tests/builds to run. Output formats include human-readable text, JSON, and CI-friendly masks. Primary target: JS/TS monorepos (pnpm/yarn/npm workspaces) and multi-language monorepos as an extension.

---

## 2) Goals & non-goals

**Goals (MVP):**

* Accurately detect changed packages between two commits (or branches).
* Build a package dependency graph (workspace-aware) for common JS/TS monorepo layouts.
* Compute the transitive closure (impact set) and produce a prioritized runlist.
* Provide a simple CLI with `--from`, `--to`, `--format` flags and a GitHub Action for CI.
* Ship as an open-source project with tests, docs, and contribution guidance.

**Non-goals (initial):**

* Full Bazel integration, patch-based binary diffs, or deep language-level analysis for all ecosystems. (These are stretch goals.)

---

## 3) Target users & use cases

* Developers and release engineers aiming to speed up CI by running only affected tests.
* Large engineering teams with monorepos who need change-impact visibility.
* OSS maintainers who want to speed up PR feedback.

Primary use cases:

* PR validation: run only impacted package tests.
* Incremental CI: reduce CI bill and latency.
* Local dev: `monodiff changed --to=HEAD` to see what to test locally.

---

## 4) MVP feature list (acceptance criteria)

1. CLI: `monodiff changed --from=main --to=HEAD --format=json`

   * Returns list of changed packages (by workspace name) and their file diffs.
2. Workspace parsers for: npm workspaces, Yarn classic workspaces, pnpm workspaces, Lerna (package.json based).
3. Dependency graph builder (reads package.json `dependencies`, `devDependencies`, `peerDependencies`) and resolves local workspace links.
4. Impact analyzer: transitive closure of affected packages.
5. Output formats: text, JSON, markdown; optionally `--ci-mask` that prints package names newline-separated for CI matrix.
6. Basic tests and CI (GitHub Actions) that run unit tests on the tool itself.

**MVP acceptance criteria:**

* Tool correctly identifies changed and impacted packages on at least three sample monorepos (include fixtures).
* Has a published GitHub repo, MIT/Apache-2.0 license, README with usage examples, CONTRIBUTING.md, and 3 `good-first-issue` labeled tasks.

---

## 5) Stretch / v2 features

* Language-specific impact detection (go modules, Python setup.py/pyproject, Maven/Gradle) via adapters.
* Support for Bazel / Buck / Nx / Rush monorepos.
* Test-selection heuristics (map changed files -> affected test files using globs or code ownership heuristics).
* Caching layer: store computed graph for a commit and reuse.
* Web UI or GitHub App showing impact visualization and historical changes.
* Integration with CI providers beyond GitHub Actions (GitLab, CircleCI, Jenkins).

---

## 6) Recommended tech stack

**Primary implementation (recommended):** Go

* Reasons: single static binary, fast startup for CLI, easy distribution to CI images, strong standard library for git and filesystem operations. Good match to Bruno's Go experience.

**Alternative:** Node.js / TypeScript

* Easier parsing of `package.json` and native integration with JS ecosystem; faster prototyping. Produce a small shim CLI for npm users.

**Libraries / tools (Go):**

* Use `go-git` or shell out to `git` for diffs (shelling to `git` keeps exact behavior).
* JSON/YAML parsing from stdlib.
* Cobra/Viper for CLI flags.

**CI / releases:**

* GitHub Actions for CI and cross-platform builds using `goreleaser` (for Go) or `pkg`/`nexe` (for Node).

---

## 7) High-level architecture

1. **Input Layer**: read CLI args (`--from`, `--to`, optional commit-range or PR number), detect repo root.
2. **Changed-file detector**: uses `git diff --name-only from..to` (or `--staged`) and normalizes paths.
3. **Workspace scanner**: discovers workspace packages by reading root package.json or config files (pnpm-workspace.yaml, lerna.json). Emits package metadata (name, path, deps).
4. **Graph builder**: constructs directed graph (package -> dependencies). Local deps are normalized to workspace names.
5. **Impact analyzer**: maps changed files -> owning package(s), then computes transitive closure by walking reverse edges in graph.
6. **Output formatter**: produces text, JSON, markdown, CI masks, and optionally a DOT graph.

Diagram (text):

```
CLI -> Git Diff -> Changed Files -> Workspace Scanner -> Package Graph -> Impact Analyzer -> Formatter -> Output
```

---

## 8) Repo layout (suggested)

```
/ (repo)
├─ cmd/monodiff/main.go        # CLI entry (cobra)
├─ internal/
│  ├─ git/                    # git-diff helpers
│  ├─ workspace/              # workspace discovery/parsers
│  ├─ graph/                  # graph builder + algorithms
│  ├─ analyzer/               # impact analysis
│  └─ output/                 # formatters
├─ fixtures/                  # sample monorepos for integration tests
├─ docs/                      # usage, design notes
├─ .github/workflows/         # CI
├─ README.md
├─ LICENSE
└─ Makefile / goreleaser.yml
```

---

## 9) Concrete TODOs (ordered, actionable)

### Milestone 0 — Project setup & discovery

* [ ] Create GitHub repo & add LICENSE (MIT/Apache-2.0), README stub, CODE\_OF\_CONDUCT, CONTRIBUTING.md.
* [ ] Add templates: ISSUE\_TEMPLATE, PULL\_REQUEST\_TEMPLATE.
* [ ] Add simple CI pipeline that runs `go test ./...` and lints.
* [ ] Add `fixtures/` with 2-3 small sample monorepos (pnpm, yarn workspace) for unit/integration tests.

### Milestone 1 — Core CLI + changed-file detection

* [ ] Implement CLI scaffold (Cobra) with commands: `changed`, `graph`, `impact`.
* [ ] Implement `git` helper module that shells out to `git` and returns changed file list for `from..to`.
* [ ] Write unit tests for `git` helper using fixtures.
* [ ] Implement `changed` command to print changed files and owning package (simple file->package mapping).
* [ ] Add acceptance tests that validate output on fixtures.

### Milestone 2 — Workspace discovery & package model

* [ ] Implement workspace parsers: detect npm workspaces, pnpm (pnpm-workspace.yaml), Yarn classic, and lerna.
* [ ] Build package metadata model: {name, path, packageJson, dependencies}
* [ ] Implement tests to verify correct package discovery on fixtures.

### Milestone 3 — Dependency graph builder

* [ ] Implement directed graph structure with helper functions: AddNode, AddEdge, ReverseDeps, Traverse.
* [ ] Parse package.json dependencies and resolve local workspace references to build edges.
* [ ] Add unit tests for graph correctness (cycle handling, missing external deps).

### Milestone 4 — Impact analysis

* [ ] Map changed files -> owning package(s) with heuristics (closest package.json ancestor, or path prefix match).
* [ ] Implement transitive closure algorithm to compute impacted packages.
* [ ] Add CLI command `impact` that lists impacted packages and includes reason (direct/indirect).
* [ ] Add JSON output format with structure: {changed:\[], impacted:\[], graph:{}}.

### Milestone 5 — CI integrations & ergonomics

* [ ] Create a GitHub Action `monodiff/action` that wraps the CLI and sets outputs for `matrix` and `affected_packages`.
* [ ] Add `--ci-mask` option to print newline-separated package names for GitHub job matrix.
* [ ] Add examples in README showing how to use the Action in PR workflows.

### Milestone 6 — Tests selection & heuristics (v1)

* [ ] Implement simple mapping from package -> test command (read from package.json `scripts.test` or a config file).
* [ ] Add `--plan` output that prints test/build commands to run for impacted packages.
* [ ] Add an option to include dependent packages with flags like `--include-dependents` or `--depth=N`.

### Milestone 7 — Performance & caching

* [ ] Add caching of package graph keyed by commit or workspace checksum.
* [ ] Add parallelism to graph traversal and file mapping for large repos.
* [ ] Add benchmark tests and perf CI job.

### Milestone 8 — Docs, publishing, community

* [ ] Write `README` with clear getting-started examples and badges.
* [ ] Add `good-first-issue` and `help-wanted` labels and 5 seeded issues.
* [ ] Publish GitHub Action marketplace listing and create initial release (use goreleaser).
* [ ] Add CODEOWNERS and a simple governance / maintainer guide.

### Milestone 9 — Extensions / v2

* [ ] Adapter plugin system for other ecosystems (Go, Python, Java). Define plugin manifest.
* [ ] Optional web UI or GitHub App hosting visualizations.
* [ ] Advanced test selection heuristics (AST-based change-to-test mapping).

---

## 10) Example CLI UX & usage

**Basic:**

```
# show changed packages between main and HEAD
monodiff changed --from=main --to=HEAD

# show impacted packages
monodiff impact --from=main --to=HEAD --format=json > impact.json

# print CI matrix mask
monodiff impact --from=main --to=HEAD --ci-mask
```

**Sample JSON output (MVP):**

```json
{
  "from": "main",
  "to": "HEAD",
  "changed": ["packages/ui", "packages/api"],
  "impacted": ["packages/ui","packages/api","packages/shared"],
  "graph": { /* minimal adjacency */ }
}
```

---

## 11) Testing strategy

* Unit tests for each internal package.
* Integration tests using `fixtures/` where we run the CLI in ephemeral git checkouts and assert outputs.
* End-to-end PR workflow test in CI using the GitHub Action and a sample repo.

---

## 12) Security & edge cases

* Handle private packages and missing `package.json` gracefully.
* Avoid executing untrusted code — only inspect manifest files and `git`.
* For very large monorepos, provide streaming output and memory-efficient graph representation.

---

## 13) Good-first-issues & contributor invites (seed ideas)

* Implement `git` helper that returns changed file list (easy).
* Add fixture: a small pnpm workspace with 3 packages and a basic test asserting changed detection.
* Implement `--format=json` output for `changed` command.

Label these as `good-first-issue` and add small `how-to` notes in CONTRIBUTING.md.

---

## 14) Release & distribution plan

* Use `goreleaser` to publish binaries for macOS, Linux, Windows and a Homebrew tap.
* Publish a GitHub Action that downloads the correct binary for the runner.
* Create initial `v0.1.0` release with CHANGELOG and a demo video/gif in README.

---

## 15) Next immediate steps (what I can do now for you)

I can immediately scaffold the GitHub-ready repo for you with:

* `README.md`, `LICENSE`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`
* CLI scaffold (`cmd/monodiff`), `internal/git` helper and a simple `changed` implementation in Go
* 2 fixtures and CI workflow to run unit tests

Tell me if you’d like the scaffold in **Go (recommended)** or **TypeScript (Node.js)** and I will generate it for you.

---

*End of plan.*
