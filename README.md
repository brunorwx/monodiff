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

1. CLI: `go run ./cmd/monodiff --from=main --to=HEAD --root=/path/to/repo --format=json`

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

*End of plan.*
