package main

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gvalid"
)

type User struct {
	// uid ＝ 属性别名 ✅（用于错误信息里显示的字段名）
	Uid   int    `v:"uid      @integer|min:1#|请输入用户ID"`
	Name  string `v:"name     @required|length:6,30#请输入用户名称|用户名称长度非法"`
	Pass1 string `v:"password1@required|password3"`
	Pass2 string `v:"password2@required|password3|same:Pass1#|密码格式不合法|两次密码不一致，请重新输入"`
}

func main() {
	// 先把数据加载到结构体，然后进行数据校验
	data2 := User{
		Uid:   1,
		Name:  "张三",
		Pass1: "123456",
		Pass2: "1234567890",
	}

	err := gvalid.New().Data(data2).Run(context.Background())
	fmt.Printf("%+v\n", err)

	// Current 方法用于获取当前层级的错误信息，通过 error 接口对象返回
	fmt.Printf("%v\n", gerror.Current(err))
}
