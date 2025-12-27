// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Book is the golang structure of table book for DAO operations like Where/Data.
type Book struct {
	g.Meta `orm:"table:book, do:true"`
	Id     any // book id
	Title  any // title
	Price  any // price
	Status any // book status
}
