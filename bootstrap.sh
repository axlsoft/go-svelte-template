#!/usr/bin/env bash
#
# bootstrap.sh <newname>
#
# Post-`gonew` re-stamp. `gonew` already rewrote the Go module path
# (github.com/OWNER/REPO -> your new module). This script handles the three
# remaining placeholders across the tree:
#
#   myapp           -> <newname>           (app name; also fixes myapp.service)
#   MYAPP_          -> <NEWNAME>_          (env var prefix)
#   myapp.service   -> <newname>.service   (systemd unit; falls out of myapp)
#
# It prints what it changed and is idempotent: running it again (or with the
# already-applied name) is a no-op.
#
# Usage:
#   gonew github.com/OWNER/REPO github.com/you/demo
#   cd demo
#   ./bootstrap.sh demo

set -euo pipefail

# --- args ------------------------------------------------------------------

if [[ $# -ne 1 || -z "${1// }" ]]; then
	echo "usage: $0 <newname>" >&2
	echo "  <newname> = lowercase app name, e.g. 'demo' or 'my-app'" >&2
	exit 2
fi

NEWNAME="$1"

# App name must be a sane, lowercase identifier (letters, digits, - and _).
if ! [[ "$NEWNAME" =~ ^[a-z][a-z0-9_-]*$ ]]; then
	echo "error: <newname> must start with a lowercase letter and contain only" >&2
	echo "       lowercase letters, digits, '-' or '_': got '$NEWNAME'" >&2
	exit 2
fi

# Env prefix: uppercase, '-' -> '_', trailing underscore. e.g. my-app -> MY_APP_
NEWPREFIX="$(printf '%s' "$NEWNAME" | tr '[:lower:]-' '[:upper:]_')_"

# --- placeholders ----------------------------------------------------------

OLD_NAME="myapp"
OLD_PREFIX="MYAPP_"

if [[ "$NEWNAME" == "$OLD_NAME" ]]; then
	echo "newname is already 'myapp'; nothing to do."
	exit 0
fi

# Resolve repo root (dir this script lives in) so paths are stable.
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

# Directories/files we must never rewrite (VCS, deps, build output, binaries).
PRUNE_DIRS=(
	"./.git"
	"./web/node_modules"
	"./web/build"
	"./internal/server/spa_dist"
)

# --- find candidate files --------------------------------------------------

# Build a `find` prune expression from PRUNE_DIRS.
find_args=(. )
first=1
for d in "${PRUNE_DIRS[@]}"; do
	if [[ $first -eq 1 ]]; then
		find_args+=( '(' -path "$d" )
		first=0
	else
		find_args+=( -o -path "$d" )
	fi
done
find_args+=( ')' -prune -o -type f -print )

changed_files=()

while IFS= read -r f; do
	# Skip this script itself to avoid self-mutation surprises.
	[[ "$f" == "./bootstrap.sh" ]] && continue
	# Skip likely-binary files (perl -i handles text only).
	if grep -Iq . "$f" 2>/dev/null; then
		if grep -Eq "$OLD_PREFIX|$OLD_NAME" "$f" 2>/dev/null; then
			before="$(cat "$f")"
			# Replace prefix first, then the app name (order independent here,
			# but prefix is uppercase so they never collide).
			perl -pi -e "s/\Q${OLD_PREFIX}\E/${NEWPREFIX}/g; s/\Q${OLD_NAME}\E/${NEWNAME}/g" "$f"
			after="$(cat "$f")"
			if [[ "$before" != "$after" ]]; then
				changed_files+=("$f")
			fi
		fi
	fi
done < <(find "${find_args[@]}")

# --- report ----------------------------------------------------------------

echo "re-stamp: app name 'myapp' -> '${NEWNAME}'"
echo "re-stamp: env prefix 'MYAPP_' -> '${NEWPREFIX}'"
echo "re-stamp: systemd unit 'myapp.service' -> '${NEWNAME}.service'"
echo

if [[ ${#changed_files[@]} -eq 0 ]]; then
	echo "no files changed (already stamped or nothing matched)."
else
	echo "changed ${#changed_files[@]} file(s):"
	for f in "${changed_files[@]}"; do
		echo "  ${f#./}"
	done
fi

echo
echo "done. review with 'git diff', then build to verify."
