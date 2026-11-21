#!/usr/bin/env bash
set -euo pipefail

if [[ $# != 0 ]]; then
    printf '用法  在生产服务器的项目目录执行 %s\n' "$0"
    [[ $# == 1 && $1 == --help ]]
    exit
fi

cd "$(dirname "$0")"
app_user=${DEPLOY_USER:-www}
for command in git go node pnpm python3 rsync flock runuser; do
    if ! command -v "$command" >/dev/null; then
        printf '缺少部署依赖 %s\n' "$command" >&2
        exit 1
    fi
done
runuser -u "$app_user" -- test -r .env
install -d -o "$app_user" bin storage storage/logs
runuser -u "$app_user" -- test -w storage
exec 9>storage/deploy.lock
if ! flock -n 9; then
    echo '已有发布正在执行' >&2
    exit 1
fi

revision=$(git rev-parse --verify HEAD)
if ! git diff --quiet HEAD --; then
    echo '项目有未提交的改动，请先确认发布内容' >&2
    exit 1
fi
source_dir=$(mktemp -d)
trap 'git worktree remove --force "$source_dir" || rmdir "$source_dir"' EXIT
git worktree add --detach "$source_dir" "$revision"
git -C "$source_dir" submodule update --init --recursive
test -f "$source_dir/docs/prismgo/zh_CN/installation.md"

(
    cd "$source_dir/admin"
    pnpm install --frozen-lockfile
    pnpm build
)
(
    cd "$source_dir/docs"
    pnpm install --frozen-lockfile
    pnpm build:all
)
(
    cd "$source_dir"
    GOWORK=off CGO_ENABLED=0 go build -mod=readonly -trimpath -o dawn-simple .
)
test "$(git rev-parse --verify HEAD)" = "$revision"
git diff --quiet HEAD --

mkdir -p docs/prismgo/zh_CN admin/dist/assets public/docs
rsync -rlt --delay-updates --delete-delay "$source_dir/docs/prismgo/zh_CN/" docs/prismgo/zh_CN/
install -o "$app_user" -m 755 "$source_dir/dawn-simple" bin/.dawn-simple.new
runuser -u "$app_user" -- ./bin/.dawn-simple.new dawn:install --force --no-interaction
runuser -u "$app_user" -- ./bin/.dawn-simple.new dawn:schema sync --force --no-interaction
mv -f bin/.dawn-simple.new bin/dawn-simple

rsync -rlt --delay-updates "$source_dir/admin/dist/assets/" admin/dist/assets/
rsync -rlt --delay-updates --exclude='/assets/' "$source_dir/admin/dist/" admin/dist/
rsync -rlt --delay-updates "$source_dir/public/docs/" public/docs/

run() {
    runuser -u "$app_user" -- ./bin/dawn-simple "$@"
}

services_ready() {
    run dawn:start "$1" status --json | python3 -c '
import json, sys
states = {service["name"]: service["state"] for service in json.load(sys.stdin)["services"]}
raise SystemExit(any(states.get(name) != "running" for name in sys.argv[1].split(",")))
' "$1"
}

if run dawn:start all status >/dev/null 2>&1; then
    run dawn:start http,ws reload
    run dawn:start cron,horizon restart
else
    runuser -u "$app_user" -- bash -c \
        'nohup ./bin/dawn-simple dawn:start all >>storage/logs/manager.log 2>&1 </dev/null &' 9>&-
fi

for ((attempt = 0; attempt < 30; attempt++)); do
    if services_ready http,ws,cron,horizon 2>/dev/null; then
        break
    fi
    sleep 1
done
if ! services_ready http,ws,cron,horizon; then
    run dawn:start all status
    exit 1
fi
printf '应用已发布 %s\n' "$revision"

run dawn:embedding install --no-interaction
run dawn:start embedding
echo '等待向量模型就绪'
for ((attempt = 0; attempt < 180; attempt++)); do
    if services_ready embedding; then
        run app:support-docs --index --no-interaction
        run dawn:start all status
        exit 0
    fi
    sleep 5
done
echo '向量模型未在15分钟内就绪，请查看服务日志' >&2
exit 1
