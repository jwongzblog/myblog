最近在忙TiDB的产品化，在集成binlog功能的时候需要利binlogctl工具执行命令来与TiDB交互，可是命令执行的返回值居然不像pd-ctl工具那样返回json格式，而是一种类似日志一样的数据格式，所以我无法使用json.Decode来优雅的还原数据的结构，返回格式如下：
```
$./binlogctl -pd-urls=http://127.0.0.1:2379 -cmd generate_meta
INFO[0000] [pd] create pd client with endpoints [http://192.168.199.118:32379]
INFO[0000] [pd] leader switches to: http://192.168.199.118:32379, previous:
INFO[0000] [pd] init cluster id 6569368151110378289
2018/06/21 11:24:47 meta.go:117: [info] meta: &{CommitTS:400962745252184065}
```

粗略来，上面的meta信息依然保持着类似**key:value**的结构形式，因此，模仿encoding/json库的实现，也能优雅的将这种日志格式还原出struct实例。

##### 第一步：拿到meta那行有效数据：
```
func GetMeta(session, scriptPath, pdAddr string) (Meta, error) {
	genCmd := fmt.Sprintf("%s/binlogctl -pd-urls=http://%s -cmd generate_meta", scriptPath, pdAddr)
	log.Debug(session, "get binlog meta",
		log.Fields{
			"scriptPath": scriptPath,
			"pdAddr":     pdAddr,
			"genCmd":     genCmd,
		})

	cmd := exec.Command("sh", "-c", genCmd)

	meta := Meta{}
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(session, "get binlog meta error", log.Fields{
			"cmd":    cmd,
			"output": output,
			"err":    err,
		})
		return meta, err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Index(line, "meta") != -1 {
			err := decode(line, &meta)
			if err != nil {
				log.Error(session, "decode meta faild", log.Fields{
					"cmd":  cmd,
					"line": line,
					"err":  err,
				})

				return meta, err
			}
		}
	}

	return meta, nil
}
```

##### 第二步：利用reflect实现decode
```
type Meta struct {
	CommitTS string
}

func decode(str string, v interface{}) error {
	vt := reflect.TypeOf(v).Elem()
	vv := reflect.ValueOf(v).Elem()

	var globalErr error
	// 遍历struct的成员
	fmt.Print(vt.NumField())
	for i := 0; i < vt.NumField(); i++ {
		field := vt.Field(i)

		// 获取字符串str中field对应的值
		value := getValue(str, field.Name)

		// 获取reflect.Value
		target := vv.FieldByName(field.Name)
		if target.Kind() != reflect.String {
			panic("Field of struct should be string(type)!")
		} else {
			// 将解析出来的值赋值给传入的interface
			target.SetString(value)
		}
	}

	defer func() {
		if err := recover(); err != nil {
			globalErr = fmt.Errorf("%v", err)
		}
	}()

	return globalErr
}
```
此处我们只处理struct成员为一层的结构且这层结构为string类型

##### 第三步：取出struct的成员名对应的值
```
func getValue(str, key string) string {
	if len(str) == 0 ||
		len(key) == 0 {
		return ""
	}

	// str:"INFO[0000] meta: &{CommitTS: 409184142902689793}"
	// index := strings.Index(str, key)
	// new := str[index :]
	// indexSplitA := strings.Index(new, ":")
	// indexSplitB := strings.Index(new, ",") || indexSplitB := strings.Index(new, "}")
	// value := new[indexSplitA + 1 : indexSplitB]
	// TrimSpace(value)
	keyIndex := strings.Index(str, key)
	if keyIndex == -1 {
		return ""
	}
	value := str[keyIndex:]

	endIndex := strings.Index(value, ",")
	if endIndex == -1 {
		endIndex = strings.Index(value, "}")
		if endIndex == -1 {
			return ""
		}
	}

	return strings.TrimSpace(value[len(key)+1 : endIndex])
}
```
此处为了节省代码使用了切片，不易读，可参考注释

如果想理解reflect为什么能让go拥有一些动态语言的特性，大家可以去看源码，我这里简单的介绍一下原理，reflect简直是go语言的后门，它将go语言的类型设计暴露了出来，咱们循着代码捋一遍：
```
reflect/type.go

// TypeOf returns the reflection Type that represents the dynamic type of i.
// If i is a nil interface value, TypeOf returns nil.
func TypeOf(i interface{}) Type {
	eface := *(*emptyInterface)(unsafe.Pointer(&i))
	return toType(eface.typ)
}
```
传进来的对象的指针被强转成emptyInterface，为什么能强转？我们看看emptyInterface是什么：
```
reflect/value.go

// emptyInterface is the header for an interface{} value.
type emptyInterface struct {
	typ  *rtype
	word unsafe.Pointer
}
```
rtype又是什么......
```
reflect/type.go

// rtype is the common implementation of most values.
// It is embedded in other struct types.
//
// rtype must be kept in sync with ../runtime/type.go:/^type._type.
type rtype struct {
	size       uintptr
	ptrdata    uintptr  // number of bytes in the type that can contain pointers
	hash       uint32   // hash of type; avoids computation in hash tables
	tflag      tflag    // extra type information flags
	align      uint8    // alignment of variable with this type
	fieldAlign uint8    // alignment of struct field with this type
	kind       uint8    // enumeration for C
	alg        *typeAlg // algorithm table
	gcdata     *byte    // garbage collection data
	str        nameOff  // string form
	ptrToThis  typeOff  // type for pointer to this type, may be zero
}
```
看这句rtype must be kept in sync with ../runtime/type.go:/^type._type
```
runtime/type.go

type _type struct {
	size       uintptr
	ptrdata    uintptr // size of memory prefix holding all pointers
	hash       uint32
	tflag      tflag
	align      uint8
	fieldalign uint8
	kind       uint8
	alg        *typeAlg
	// gcdata stores the GC type data for the garbage collector.
	// If the KindGCProg bit is set in kind, gcdata is a GC program.
	// Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
	gcdata    *byte
	str       nameOff
	ptrToThis typeOff
}
```
本质上rtype和_type的内存布局是一样的，所以他们能够通过强制类型转换，将基本类型runtime._type转成reflect.rtype，这样就具备了_type对成员变量的操作能力，与python的object（万物皆对象）设计有异曲同工之妙