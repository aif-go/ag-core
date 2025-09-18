#!/bin/bash
kitex_version="v0.14.1"
kdir="kitex_${kitex_version}"
diffdir="kitex_diff"

if [ ! -d $diffdir ]; then
    mkdir $diffdir
fi

# 从kitex仓库clone代码
echo "# clone kitex code ${kitex_version}"
git clone --depth 1 --branch $kitex_version git@github.com:cloudwego/kitex.git $kdir

# 生成diff文件
echo "# generate diff file"
diff -u ./${kdir}/tool/internal_pkg/tpl/service.go service.go > ${diffdir}/service.go.diff
diff -u ./${kdir}/tool/internal_pkg/tpl/server.go  server.go  > ${diffdir}/server.go.diff
diff -u ./${kdir}/tool/internal_pkg/tpl/client.go  client.go  > ${diffdir}/client.go.diff

# 清理kitex代码
echo "# clean kitex code"
rm -rf $kdir
