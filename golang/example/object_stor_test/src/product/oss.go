package product

import (
	"log"
	"os"
	"src/common"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Oss struct {
	path, objName string
}

func NewOss(path, objName string) *Oss {
	return &Oss{path: path, objName: objName}
}

func (o *Oss) Upload() error {
	// 创建OSSClient实例。
	client, err := oss.New(common.Config.FileHost, common.Config.PublicKey, common.Config.PrivateKey)
	if err != nil {
		return err
	}

	bucketName := common.Config.BucketName
	objectName := o.objName
	locaFilename := o.path

	// 获取存储空间。
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return err
	}
	chunks, err := oss.SplitFileByPartNum(locaFilename, 10)
	if err != nil {
		return err
	}
	fd, err := os.Open(locaFilename)
	if err != nil {
		return err
	}
	defer fd.Close()

	// 指定存储类型为标准存储。
	storageType := oss.ObjectStorageClass(oss.StorageStandard)

	// 步骤1：初始化一个分片上传事件，并指定存储类型为标准存储。
	imur, err := bucket.InitiateMultipartUpload(objectName, storageType)
	// 步骤2：上传分片。
	var (
		parts []oss.UploadPart
	)

	for _, chunk := range chunks {
		fd.Seek(chunk.Offset, os.SEEK_SET)
		// 调用UploadPart方法上传每个分片。
		part, err := bucket.UploadPart(imur, fd, chunk.Size, chunk.Number)
		if err != nil {
			log.Println(err)
			return err
		}
		parts = append(parts, part)
	}

	// 指定Object的读写权限为公共读，默认为继承Bucket的读写权限。
	objectAcl := oss.ObjectACL(oss.ACLPublicRead)

	// 步骤3：完成分片上传，指定文件读写权限为公共读。
	cmur, err := bucket.CompleteMultipartUpload(imur, parts, objectAcl)
	if err != nil {
		return err
	}
	log.Println(cmur)

	return nil
}
