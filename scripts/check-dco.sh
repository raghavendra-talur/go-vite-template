#!/usr/bin/env bash
set -euo pipefail

if [[ "$#" -ne 2 ]]; then
  echo "usage: scripts/check-dco.sh <base-ref> <head-ref>" >&2
  exit 1
fi

base_ref="$1"
head_ref="$2"

mapfile -t commits < <(git rev-list --reverse --no-merges "${base_ref}..${head_ref}")

if [[ "${#commits[@]}" -eq 0 ]]; then
  echo "No non-merge commits to check."
  exit 0
fi

echo "Checking DCO sign-offs for ${#commits[@]} commit(s)..."

failed=0

for commit in "${commits[@]}"; do
  subject="$(git show -s --format=%s "$commit")"
  author_name="$(git show -s --format=%an "$commit")"
  author_email="$(git show -s --format=%ae "$commit")"
  expected="Signed-off-by: ${author_name} <${author_email}>"
  trailers="$(git show -s --format=%B "$commit" | git interpret-trailers --parse)"

  if printf '%s\n' "$trailers" | awk -v expected="$expected" 'tolower($0) == tolower(expected) { found = 1 } END { exit(found ? 0 : 1) }'; then
    echo "PASS  $commit  $subject"
  else
    failed=1
    echo "::error title=Missing DCO sign-off::$commit ($subject) is missing '${expected}'"
  fi
done

if [[ "$failed" -ne 0 ]]; then
  cat <<'EOF'
Every contributed commit must include a matching Signed-off-by trailer.

Add one for your latest commit with:
  git commit --amend --signoff

Add sign-offs across a branch with:
  git rebase --signoff <base>
EOF
  exit 1
fi
