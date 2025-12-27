// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Book is the golang structure for table book.
type Book struct {
	Id     uint   `json:"id"     orm:"id"     description:"book id"`     // book id
	Title  string `json:"title"  orm:"title"  description:"title"`       // title
	Price  uint   `json:"price"  orm:"price"  description:"price"`       // price
	Status int    `json:"status" orm:"status" description:"book status"` // book status
}
