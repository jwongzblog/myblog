在工程里面，我们喜欢使用map来缓存一些值，但是如果要修改这些值，需要注意一点细节

value如果是object，则map不允许直接修改这个值，或者通过`val,ok := mapA[key]`获取val修改的，只是修改一个浅拷贝的新值，如
```
package main

import "fmt"

type Foo struct {
    Bar int64
}

func main() {
	test := make(map[string]Foo)
	test["a"] = Foo{Bar:123}
	
	fmt.Print(test)
	fmt.Print("\n")
	
	val,ok := test["a"]
	fmt.Print(val)
	fmt.Print("\n")
	fmt.Print(ok)
	fmt.Print("\n")
	
	val.Bar = 456
	fmt.Print(test["a"].Bar)

	// test["a"].Bar = 789 //error:cannot assign to struct field test["a"].Bar in map
}

output:
map[a:{123}]
{123}
true
123
Program exited.
```

如果value是指针，则可以被修改
```
package main

import "fmt"

type Foo struct {
    Bar int64
}

func main() {
	test := make(map[string]*Foo)
	test["a"] = &Foo{Bar:123}
	
	fmt.Print(test)
	fmt.Print("\n")
	
	val,ok := test["a"]
	fmt.Print(val)
	fmt.Print("\n")
	fmt.Print(ok)
	fmt.Print("\n")
	
	val.Bar = 456
	fmt.Print(test["a"].Bar)

	fmt.Print("\n")
	test["a"].Bar = 789
	fmt.Print(test["a"].Bar)
}

output:
map[a:0x40e020]
&{123}
true
456
789
Program exited.
```