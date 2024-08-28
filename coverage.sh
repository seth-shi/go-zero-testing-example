#!/bin/bash

find_go_mod() {
    local dir="$1"
    local depth="$2"

    if [ $depth -gt 3 ]; then
        echo ""
        return
    fi

    if [ -f "$dir/go.mod" ]; then
        echo "$dir"
        return
    fi

    new_dir=$(dirname "$dir")
    result=$(find_go_mod "$new_dir" $((depth + 1)))
    echo "$result"
}

current_dir=$(pwd)
found_path=$(find_go_mod "$current_dir" 0)

if [ -z "$found_path" ]; then
    echo "未找到 go.mod 文件，退出"
    exit 1
fi


## 测试的包
dir=$1
if [ -z "$dir" ]; then
    dir="..."
else
    dir="$dir"
fi
# 测试覆盖率
go test "$found_path/$dir" -coverprofile=coverage.cov.tmp -covermode count
## 等效 Gitlab 的 -ignore-gen-files
cat coverage.cov.tmp | grep -wv -e gen.go -e pb.go -e internal/server -e internal/faker -e internal/mock > coverage.out
rm coverage.cov.tmp
go tool cover -html coverage.out
