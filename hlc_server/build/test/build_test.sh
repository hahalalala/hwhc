env="test"

coderepo=${GOPATH}"/src/github.com/hwhc"  #代码仓库
rundir=${GOPATH}"/src/github.com/hwhcrun" #运行文件夹
bakdir=${GOPATH}"/src/github.com/hwhcrun" #备份文件夹

#拉代码
cd ${coderepo}
echo "start pull code "
git checkout main
git pull
echo "pull code finish"

now=$(date "+%Y%m%d%H%M%S")
#备份上一次
mv ${rundir}/${env}walle ${bakdir}/${env}walle${now}

#创建新的运行文件夹
mkdir ${rundir}/${env}walle/
cp ${coderepo}/hlc_server/build/${env}/conf.json  ${rundir}/${env}walle/

echo "start build..."
go build -o  ${rundir}/${env}walle/hlcmain ${coderepo}/hlc_server/main.go
echo "build finish"

#重启服务
/bin/systemctl restart supervisord


echo "deploy success"