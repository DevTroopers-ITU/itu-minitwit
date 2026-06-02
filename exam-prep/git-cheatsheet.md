# Exam terminal cheatsheet — proving claims with git

Run these live if an examiner says "show me." Keep findings.md open alongside.

## Core commands

| Goal | Command |
|------|---------|
| When did a string enter/leave the code? (pickaxe) | `git log -S "TEXT" --oneline --reverse --pretty='%h %ad %s' --date=short` |
| Show one commit's full diff | `git show <hash>` |
| Full history of ONE function | `git log -L :funcName:file.go` |
| Who/when wrote a line | `git blame file.go` (or `git blame -L 123,130 file.go`) |
| Prove current code state | `git grep -n "TEXT" -- '*.go'` |
| File history through renames | `git log --follow --oneline -- path/file` |
| All commits touching a file | `git log --oneline -- path/file` |

## Ready-made proofs of our findings

```fish
# Secret-key fallback: when + the exact change
git log -S "dev-fallback-key-change-in-production" --oneline
git show 2acff4c                       # "make key pass tests" added the fallback
git log -L :getSecretKey:main.go       # whole evolution of the function

# Secret-key timeline (security fix -> quick patch -> swarm migration)
git log -S "godotenv" --oneline --reverse        # 57a4f1c removed hardcoded key
git log -S "/run/secrets" --oneline --reverse    # 3c649da swarm added secret files

# SQLite is test-only; prod is Postgres
git grep -n "driver/sqlite" -- '*.go'    # only *_test.go
git grep -n "driver/postgres" -- '*.go'  # db.go (prod)

# Browser test exists and runs in CI (proves progress.md S07 is stale)
git grep -n "minitwit_ui_browser_test" .github/

# How prod was really provisioned (manual, not Terraform)
git log --oneline -- terraform/          # only ~4 commits, added late
grep -n "already provisioned" docs/operations/docker-swarm.md

# Test-file evolution through the Python->Go rename
git log --follow --oneline -- python-references/minitwit_tests.py
```

## Live prod check (needs ssh as root@64.226.116.162 — ask before doing it)
```fish
ssh root@64.226.116.162 'id=$(docker ps -q -f name=minitwit_webserver | head -1); docker exec $id sh -c "test -f /app/.env && echo HAS_DOTENV || echo NO_DOTENV; ls /run/secrets"'
```

Tip: `--pretty='%h %ad %s' --date=short` on any `git log` adds short hash + date + subject.
