package entities

type ProductDiscountGroup struct {
	StudentProduct StudentProduct
	ProductGroups  []ProductGroup
	DiscountType   string
}
