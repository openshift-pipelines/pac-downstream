# Pipelines-as-Code AI Review Rules

Paco uses these project-specific rules in addition to checking for concrete
bugs, security issues, and missed edge cases.

## Review Priorities

- Report only issues introduced or exposed by the diff. Do not comment on
  unchanged code unless the change makes an existing defect reachable.
- Prefer correctness, security, data loss, compatibility, and concurrency
  findings over style. Explain the concrete failure mode, not just the rule.
- Do not report formatting, import ordering, spelling, or other issues that an
  existing formatter or linter will catch.
- Avoid speculative findings. If a claim depends on repository context absent
  from the diff, do not present it as a defect.
- Report one comment per root cause. Do not repeat the same issue at every
  affected call site.
- Use `critical` only for exploitable vulnerabilities or likely data loss,
  `high` for definite correctness failures, `medium` for bounded bugs or
  important missing coverage, and `low` for objective project-rule violations.

## Testing

- Use table-driven tests with an anonymous struct slice
  (`tests := []struct{...}{...}`) iterated with
  `for _, tt := range tests { t.Run(tt.name, ...) }`.
- No underscores in test function names; use PascalCase
  (`TestGetTektonDir`, not `TestGetTektonDir_Something`).
- Give each table-driven case a descriptive `name` field used in `t.Run`.
- For complex per-case setup, add a `setup func(t *testing.T, ...)` field to
  the test struct instead of writing a separate test function.
- Use `gotest.tools/v3/assert` for assertions — never `testify` or a custom
  `pkg/assert` package. Prefer `assert.NilError`, `assert.Assert`,
  `assert.Equal`, `assert.ErrorContains` over manual `if err != nil` checks.
- E2E test names should be descriptive and provider-specific where
  applicable; avoid generic names like `TestRun`/`TestProcess`.
- Flag missing tests only when the changed behavior has a meaningful failure
  path or regression risk. Do not request tests for mechanical refactors.

## Error Handling

- Wrap errors with `fmt.Errorf("...: %w", err)` when crossing an abstraction
  boundary or when the added context helps identify the failed operation.
  Preserve errors unchanged when callers rely on their exact identity.
- Prefer `errors.Is`/typed sentinel errors for known domain error checks
  instead of string matching on error messages.
- Never use `panic` in production code paths (test helpers/stubs are
  exempt).
- Do not silently convert required-operation failures into success. Optional
  enrichment may continue after logging a clear warning.

## Style / Go Idioms

- Provider code (`pkg/provider/`) must log through the provider's own
  logger (`v.Logger`), never `run.Clients.Log` — the provider logger
  carries provider/repository/event-id context that the global logger
  lacks. This applies to GitHub, GitLab, Bitbucket, Bitbucket Data Center,
  and Gitea.
- Keep `context.Context` as the first parameter for functions that need
  one.
- Avoid `init()` functions outside of already-exempted packages.
- Prefer explicit returns over naked returns.
- Keep logging structured and logger-safe; don't bypass the configured
  logger with ad-hoc `fmt.Println`/`log.Print` in production code.

## Shell and Tekton

- Quote variable expansions and use `set -euo pipefail` in non-trivial Bash
  scripts unless a documented compatibility constraint prevents it.
- Pass large or untrusted content through files, standard input, or
  environment variables rather than interpolating it into shell source or a
  single command-line argument.
- Put explicit timeouts around external services and long-running model or
  network calls.
- Keep required API failures visible. If a best-effort API call is allowed to
  fail, initialize its output to a valid, conservative value.
- Never print raw model output, API errors, diffs, or environment state when
  they may contain credentials. Scrub diagnostic output before logging it.

## Documentation

- User documentation lives under `docs/content/docs/` (Hugo + hugo-book
  theme). New or changed user-facing behavior should include a matching doc
  update when users need new instructions or reference material.
- Follow existing frontmatter and structure when adding new doc pages; use
  the repo's custom shortcodes (`tech_preview`, `support_matrix`) where
  appropriate instead of ad-hoc formatting.

## Security

- Webhook-driven triggers must be authenticated: GitHub App requests via
  JWT/signature verification, webhook-based providers via a shared webhook
  secret. Flag any change that weakens or bypasses this.
- Secrets (GitHub App private key, application ID, webhook secret, tokens)
  must be sourced from Kubernetes Secrets, never hardcoded or logged in
  plaintext.
- Incoming webhook triggers must require the configured shared secret and
  explicit branch targeting — flag anything that could allow
  unauthenticated or ambiguously-targeted triggers.
- Be careful with changes to GitHub token scoping — token permissions
  should stay minimal and fail closed when a repository pattern doesn't
  match.
- Changes that process user-controlled input, including GitOps and pull request
  comments, must assess whether that input is sanitized before it can be
  interpreted or executed.

## Dependencies and Generated Files

- Changes to Go dependencies must keep `go.mod`, `go.sum`, and `vendor/`
  consistent.
- When a dependency is added or updated, check the specific version against an
  online security/vulnerability database (GitHub Advisory Database, OSV, CVE) to
  determine whether it has been compromised, yanked, or has a known
  vulnerability. If the dependency is compromised or malicious, report it at
  `critical` severity and prefix the comment with **SUPER HIGH SECURITY** in
  bold, then explain the compromise and the safe version to move to.
- Do not ask for hand edits to generated files. Check that the source of truth
  changed and that corresponding generated output was refreshed.
