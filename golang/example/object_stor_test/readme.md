# build
```
export GOPATH="$PWD/object_stor_test"
```

```
cd src
go build main.go
```

# configure
```
{
    "说明1":"us3:管理 bucket 创建和删除必须要公私钥，如果只做文件上传和下载用TOEKN就够了，为了安全，强烈建议只使用 TOKEN 做文件管理",
    "说明2":"oss:access_id、access_key",
    "public_key":"",
    "private_key":"",

    "说明3":"以下两个参数是用来管理文件用的。对应的是 file.go 里面的接口，file_host 是不带 bucket 名字的。比如：北京地域的host填cn-bj.ufileos.com，而不是填 bucketname.cn-bj.ufileos.com。如果是自定义域名，请直接带上 http 开头的 URL。如：http://example.com，而不是填 example.com。",
    "说明4":"oss endpoint",
    "bucket_name":"",
    "file_host":"",

    "说明5":"以下参数是用来管理 bucket 用的。对应的是 bucket.go 里面的接口",
    "bucket_host":""
}
```