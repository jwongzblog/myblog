package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	dir = flag.String("dir", "/mnt/d/test_dir", `target directory:
/mnt/d/test_dir(linux)
d:\test_dir(windows)`)
	totalSize = flag.Int64("totalSize", 10, "total size : GB")
	fileSize  = flag.String("fileSize", "32MB", `4KB
64KB
32MB
64MB
300MB
1GB
5GB`)

	sizeMap = map[string]int64{
		"4KB":   1024 * 4,
		"64KB":  1024 * 64,
		"32MB":  1024 * 1024 * 32,
		"64MB":  1024 * 1024 * 64,
		"300MB": 1024 * 1024 * 300,
		"1GB":   1024 * 1024 * 1024,
		"5GB":   1024 * 1024 * 1024 * 5,
	}
)

func generatorFile(path string, totalByte, fileByte int64) {
	err := os.Mkdir(path, 0777)
	if os.IsNotExist(err) {
		log.Printf("create directory %s:%s", path, err.Error())
		return
	}

	fileCount := totalByte / fileByte
	for i := int64(0); i < fileCount; i++ {
		fileName := filepath.Join(path, fmt.Sprint(i))
		f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			log.Print(err)
		}
		defer f.Close()

		writeCount := fileByte / 1024 / 4
		for j := int64(0); j < writeCount; j++ {
			randomByte := make([]byte, 1024*4)
			rand.Read(randomByte)
			_, err := f.Write(randomByte)
			if err != nil {
				log.Print(err)
			}
		}
		err = f.Sync()
		if err != nil {
			log.Print(err)
		}
	}
}

func generator() {
	totalByte := *totalSize * 1024 * 1024 * 1024
	for k, v := range sizeMap {
		if v <= totalByte && strings.EqualFold(*fileSize, k) {
			generatorFile(filepath.Join(*dir, k), totalByte, v)
			break
		}
	}
}

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Parse()

	if !strings.EqualFold(runtime.GOOS, "linux") {
		*dir = `d:\test_dir`
	}

	err := os.Mkdir(*dir, 0777)
	if os.IsNotExist(err) {
		log.Println("target directory:", err.Error())
		return
	}

	generator()
}
