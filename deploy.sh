#!/usr/bin/env bash
set -euo pipefail

if [[ $# != 2 || ${1:-} == --help ]]; then
    printf '用法  %s <SSH目标> <线上项目绝对路径>\n' "$0"
    [[ ${1:-} == --help ]]
    exit
fi

cd "$(dirname "$0")"
git submodule update --init --recursive
test -f docs/prismgo/zh_CN/installation.md
npm --prefix tools ci
pnpm --dir admin install --frozen-lockfile
pnpm --dir docs install --frozen-lockfile
pnpm --dir admin build
pnpm --dir docs build:all

release=$(mktemp -d)
trap 'rm -rf "$release"' EXIT
GOWORK=off GOOS=linux GOARCH="${DEPLOY_ARCH:-amd64}" CGO_ENABLED=0 \
    go build -trimpath -o "$release/dawn-simple" .

remote_script=$(cat <<'REMOTE'
set -euo pipefail
cd "$1"
app_user=$2
test -f .env
mkdir -p storage
exec 9>storage/deploy.lock
flock -n 9
release=$(mktemp -d storage/.deploy.XXXXXX)
trap 'rm -rf "$release"' EXIT
tar -xf - -C "$release"
install -d -o "$app_user" bin storage/logs
runuser -u "$app_user" -- test -w storage

run() {
    runuser -u "$app_user" -- ./bin/dawn-simple "$@"
}

mkdir -p database docs/prismgo/zh_CN
rsync -rlt --delay-updates "$release/database/" database/
rsync -rlt --delay-updates --delete --include='/*.md' --exclude='*' "$release/docs/" docs/
rsync -rlt --delay-updates --delete "$release/docs/prismgo/zh_CN/" docs/prismgo/zh_CN/
install -o "$app_user" -m 755 "$release/dawn-simple" bin/.dawn-simple.new
runuser -u "$app_user" -- ./bin/.dawn-simple.new dawn:install --force --no-interaction
runuser -u "$app_user" -- ./bin/.dawn-simple.new dawn:schema sync --force --no-interaction
mv -f bin/.dawn-simple.new bin/dawn-simple

for path in admin/dist public/docs; do
    mkdir -p "$path"
    rsync -rlt --delay-updates "$release/$path/" "$path/"
done

if run dawn:start all status >/dev/null 2>&1; then
    run dawn:start http,ws reload
    run dawn:start cron,horizon restart
else
    runuser -u "$app_user" -- bash -c \
        'nohup ./bin/dawn-simple dawn:start all >>storage/logs/manager.log 2>&1 </dev/null &' 9>&-
    for ((attempt = 0; attempt < 30; attempt++)); do
        if run dawn:start all status >/dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    run dawn:start all status
fi

run dawn:embedding install --no-interaction
run dawn:start embedding
echo '等待向量模型就绪'
for ((attempt = 0; attempt < 180; attempt++)); do
    if run dawn:start embedding status --json | python3 -c \
        'import json, sys; raise SystemExit(json.load(sys.stdin)["services"][0]["state"] != "running")'; then
        run app:support-docs --index --no-interaction
        run dawn:start all status
        exit 0
    fi
    sleep 5
done
echo '向量模型未在15分钟内就绪，请查看服务日志' >&2
exit 1
REMOTE
)
remote_command=$(printf 'bash -c %q -- %q %q' "$remote_script" "$2" "${DEPLOY_USER:-www}")
COPYFILE_DISABLE=1 tar --no-xattrs -cf - -C "$release" dawn-simple -C "$PWD" admin/dist public/docs database docs/*.md docs/prismgo/zh_CN \
    | ssh -o ServerAliveInterval=15 -o ServerAliveCountMax=4 "$1" "$remote_command"
