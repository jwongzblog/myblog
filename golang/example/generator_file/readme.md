# build
go build file_generator.go

# exec
use default args
```
./file_generator
./file_generator.exe
```

for more args
```
./file_generator --help
Usage of ./file_generator:
  -dir string
    	target directory:
    	/mnt/d/test_dir(linux)
    	d:\test_dir(windows) (default "/mnt/d/test_dir")
  -fileSize string
    	4KB
    	64KB
    	32MB
    	64MB
    	300MB
    	1GB
    	5GB (default "32MB")
  -totalSize int
    	total size : GB (default 10)
```