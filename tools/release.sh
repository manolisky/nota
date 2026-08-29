#!/bin/sh
# Bump nota.sty, commit, and tag. Usage: make release bump=patch|minor|major
set -eu

ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
STY="$ROOT/latex/nota.sty"

normalize_version() {
	v=${1#v}
	case $v in
	*.*.*) printf '%s\n' "$v" ;;
	*.*) printf '%s.0\n' "$v" ;;
	*) printf '%s.0.0\n' "$v" ;;
	esac
}

bump_version() {
	kind=$1
	cur=$(normalize_version "$2")
	old_ifs=$IFS
	IFS=.
	# shellcheck disable=SC2086
	set -- $cur
	IFS=$old_ifs
	major=$1
	minor=$2
	patch=$3
	case $kind in
	patch) patch=$((patch + 1)) ;;
	minor)
		minor=$((minor + 1))
		patch=0
		;;
	major)
		major=$((major + 1))
		minor=0
		patch=0
		;;
	*)
		echo "release: bump must be patch, minor, or major (got '$kind')" >&2
		return 1
		;;
	esac
	printf '%s.%s.%s\n' "$major" "$minor" "$patch"
}

sty_version() {
	sed -n 's/.*ProvidesPackage{nota}\[[^]]* v\([0-9][0-9.]*\) .*/\1/p' "$STY"
}

latest_tag_version() {
	git -C "$ROOT" tag -l 'v*' --sort=-v:refname 2>/dev/null | head -n 1 | sed 's/^v//'
}

current_version() {
	tag=$(latest_tag_version)
	if [ -n "$tag" ]; then
		printf '%s\n' "$tag"
		return
	fi
	sty=$(sty_version)
	if [ -z "$sty" ]; then
		echo "release: no git tag and could not read version from $STY" >&2
		return 1
	fi
	printf '%s\n' "$sty"
}

sty_version_matches_tag() {
	tag=${1#v}
	sty=$(sty_version)
	sty_n=$(normalize_version "$sty")
	tag_n=$(normalize_version "$tag")
	if [ "$sty_n" != "$tag_n" ]; then
		echo "release: latex/nota.sty version $sty does not match tag v$tag" >&2
		return 1
	fi
}

write_sty_version() {
	next=$1
	today=$(date +%Y/%m/%d)
	tmp=$(mktemp)
	awk -v today="$today" -v ver="$next" '
		/ProvidesPackage\{nota\}/ {
			printf "\\ProvidesPackage{nota}[%s v%s Nota Package]\n", today, ver
			next
		}
		{ print }
	' "$STY" >"$tmp"
	mv "$tmp" "$STY"
}

self_test() {
	fail=0
	check() {
		got=$1
		want=$2
		msg=$3
		if [ "$got" != "$want" ]; then
			echo "self-test fail: $msg (got '$got' want '$want')" >&2
			fail=1
		fi
	}
	check "$(normalize_version 1.2)" "1.2.0" "pad two-part version"
	check "$(normalize_version v1.2.0)" "1.2.0" "strip v prefix"
	check "$(bump_version patch 1.2.0)" "1.2.1" "patch"
	check "$(bump_version minor 1.2.0)" "1.3.0" "minor"
	check "$(bump_version major 1.2.0)" "2.0.0" "major"
	check "$(bump_version patch 1.2)" "1.2.1" "patch from two-part"
	if [ "$fail" -ne 0 ]; then
		exit 1
	fi
	echo "release.sh self-test ok"
}

cmd=${1:-}

if [ "$cmd" = "--self-test" ]; then
	self_test
	exit 0
fi

if [ "$cmd" = "check-tag" ]; then
	tag=${2:-}
	if [ -z "$tag" ]; then
		echo "release: usage: $0 check-tag vX.Y.Z" >&2
		exit 1
	fi
	sty_version_matches_tag "$tag"
	echo "release: nota.sty matches $tag"
	exit 0
fi

bump=$cmd
if [ -z "$bump" ]; then
	echo "release: missing bump. Use: make release bump=patch   (or minor, major)" >&2
	exit 1
fi
case $bump in
patch | minor | major) ;;
*)
	echo "release: bump must be patch, minor, or major (got '$bump')" >&2
	exit 1
	;;
esac

cd "$ROOT"

if [ -n "$(git status --porcelain)" ]; then
	echo "release: working tree is dirty; commit or stash first" >&2
	git status --porcelain >&2
	exit 1
fi

cur=$(current_version)
next=$(bump_version "$bump" "$cur")
tag=v$next

if git rev-parse "$tag" >/dev/null 2>&1; then
	echo "release: tag $tag already exists" >&2
	exit 1
fi

write_sty_version "$next"
git add latex/nota.sty
git commit -m "Release $tag"
git tag -a "$tag" -m "Release $tag"

echo "Created commit and tag $tag (was $(normalize_version "$cur"))."
echo "Push with:"
echo "  git push origin main --tags"
