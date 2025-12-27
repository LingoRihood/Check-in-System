package v1

// book 相关的增删改查的请求参数结构体和返回响应的结构体

import (
	"gf_demo/internal/model/entity"

	"github.com/gogf/gf/v2/frame/g"
)

// CreateReq 创建书籍请求结构体
type CreateReq struct {
	g.Meta `path:"/books" method:"post" tags:"book" summary:"创建书籍"`
	Title  string `p:"title" v:"required|length:1,45#书名必须填写|书名长度必须满足1-45字" dc:"书籍标题"`
	Price  int    `p:"price" v:"required|min:1#价格必须填写|价格不能小于1" dc:"书籍价格"`
}

// CreateRes 创建书籍响应结构体
type CreateRes struct {
	// mime 是 MIME Type（媒体类型） 的简称，用来告诉客户端“这段数据是什么格式”
	g.Meta `mime:"application/json"`

	// json:"id"：序列化成 JSON 时字段名用小写 id, 字段类型：64 位整数。
	Id int64 `json:"id" dc:"book id"`
}

// DeleteReq 删除书籍请求结构体
type DeleteReq struct {
	g.Meta `path:"/books/{id}" method:"delete" tags:"book" summary:"删除书籍"`
	Id     int64 `p:"id" v:"required|min:1#ID必须传|ID不能小于1" dc:"书籍ID"`
}

// DeleteRes 删除书籍响应结构体
type DeleteRes struct{}

// Status 书籍状态枚举
type Status = int8

const (
	StatusAvailable Status = 0 // 上架
	StatusDisable   Status = 1 // 下架
)

// UpdateReq 更新书籍请求结构体
type UpdateReq struct {
	// 这表示 API 请求的路径（路由）。books/{id} 是一个包含路径参数的 URL 模板，{id} 是路径中的占位符，表示书籍的唯一 ID。
	// 例如，当你访问 PUT /books/123 时，123 就是书籍的 ID
	g.Meta `path:"/books/{id}" method:"put" tags:"book" summary:"更新书籍"`
	Id     int64  `p:"id" v:"required|min:1#ID必须传|ID不能小于1" dc:"书籍ID"`
	Title  string `p:"title" v:"length:1,45#书名长度必须满足1-45字" dc:"书籍标题"`
	Price  int    `p:"price" v:"min:1#价格不能小于1" dc:"书籍价格"`
	Status Status `p:"status" v:"in:0,1#状态只能是上架或下架" dc:"书籍状态"`
}

// UpdateRes 更新书籍响应结构体
type UpdateRes struct{}

// GetOneReq 获取单个用户请求结构体
type GetOneReq struct {
	g.Meta `path:"/books/{id}" method:"get" tags:"book" summary:"获取书籍"`
	Id     int64 `p:"id" v:"required|min:1#ID必须传|ID不能小于1" dc:"书籍ID"`
}

type GetOneRes struct {
	// 使用生成的 model 结构体
	*entity.Book `json:"book"`
}

// GetListReq 获取书籍列表请求结构体
type GetListReq struct {
	g.Meta `path:"/books" method:"get" tags:"book" summary:"获取书籍列表"`
	Status Status `p:"status" v:"required|in:0,1#status必传|状态只能是上架或下架" dc:"书籍状态"`
}

type GetListRes struct {
	// 使用生成的 model 结构体
	List []*entity.Book `json:"list" dc:"book list"`
}
