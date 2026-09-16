#!/usr/bin/env bash
set -euo pipefail

version="v$1"
repo="${INFRA_REPO:-leorolland/infra}"
base="${INFRA_BASE_BRANCH:-}"
clone_url="${INFRA_CLONE_URL:-https://x-access-token:${GH_TOKEN}@github.com/${repo}.git}"
branch="sortir/bump-${version}"
defaults="roles/sortir/defaults/main.yml"

workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT

git clone --quiet "$clone_url" "$workdir"
cd "$workdir"
git config user.name "github-actions[bot]"
git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
if [ -n "$base" ]; then
  git checkout --quiet -B "$branch" "origin/$base"
else
  git checkout --quiet -B "$branch"
fi

sed -i.bak "s|^sortir_version:.*|sortir_version: \"${version}\"|" "$defaults"
sed -i.bak 's|^sortir_arch:.*|sortir_arch: "amd64"|' "$defaults"
rm -f "$defaults.bak"

if git diff --quiet; then
  echo "infra already targets ${version}, nothing to do"
  exit 0
fi

git add "$defaults"
git commit --quiet -m "chore(sortir): bump to ${version}"
git push --quiet --force-with-lease -u origin "$branch"

open_pr=$(gh pr list --repo "$repo" --head "$branch" --state open --json url --jq '.[0].url // empty')
if [ -n "$open_pr" ]; then
  echo "PR already open: $open_pr"
  exit 0
fi

pr_args=(--repo "$repo" --head "$branch" --title "sortir.in ${version}" --body "Bump [sortir.in](https://github.com/leorolland/sortir.in) to [\`${version}\`](https://github.com/leorolland/sortir.in/releases/tag/${version}).

Merge, then run the \`sortir\` playbook to deploy.

> \`sortir_arch\` is converged to \`amd64\` to match the release binaries.")
if [ -n "$base" ]; then
  pr_args+=(--base "$base")
fi

gh pr create "${pr_args[@]}"
