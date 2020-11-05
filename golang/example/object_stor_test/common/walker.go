package common

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetFileMap(filePath string) (map[string]string, int64, error) {
	var (
		baseDir string
		size int64
	)
	fileMap := make(map[string]string)

	// 移除末尾字符"/"
	filePath = strings.TrimRight(filePath, "/")
	// 从filePath开始深度优先遍历
	err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 构造bucket中的路径
		// remoteFileKey经过trim后为空，或者为"/xxx_xxxx.sst"
		remoteFileKey := strings.TrimPrefix(path, filePath)
		if len(remoteFileKey) == 0 {
			// 如果路径完全匹配，一定是根目录或者是文件类型
			remoteFileKey = info.Name()
			// 此次所有上传bucket文件的一级目录
			// baseDir = "backup_20200923"
			baseDir = info.Name()
		} else {
			// 拼接成"backup_20200923/xxx_xxxx.sst"
			remoteFileKey = fmt.Sprintf("%s%s", baseDir, remoteFileKey)
		}

		// 如果是目录，跳过
		if info.IsDir() {
			return nil
		}
		size += info.Size()

		fileMap[path] = remoteFileKey

		return nil
	})

	if err != nil {
		return fileMap, size, err
	}

	return fileMap, size, nil
}
