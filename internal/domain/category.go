package domain

/*
Transaction categories. These are the only values a Transaction's
Category may hold. Every Categorizer implementation must return one of
them, and the usecase falls back to CategoryUncategorized when it gets
anything else, so the list lives here once instead of being repeated
in each implementation.
*/
const (
	CategoryFood           = "food"
	CategoryTransportation = "transportation"
	CategoryShopping       = "shopping"
	CategoryBills          = "bills"
	CategoryEntertainment  = "entertainment"
	CategoryHealth         = "health"

	/*
		CategoryOthers is a real pick: the categorizer understood the
		transaction and it fits none of the categories above (rent,
		donations, transfers, ...). It differs from
		CategoryUncategorized, which means no category could be
		determined at all.
	*/
	CategoryOthers = "others"

	/*
		CategoryUncategorized is used when no category could be
		determined: nothing matched, the categorizer was unsure, or it
		failed. It is valid to store but is never a categorizer's pick.
	*/
	CategoryUncategorized = "uncategorized"
)

/*
Categories lists every category a categorizer may choose, in display
order. It excludes CategoryUncategorized, which is a fallback rather
than a choice.
*/
var Categories = []string{
	CategoryFood,
	CategoryTransportation,
	CategoryShopping,
	CategoryBills,
	CategoryEntertainment,
	CategoryHealth,
	CategoryOthers,
}

/*
IsValidCategory reports whether category is one of Categories or
CategoryUncategorized.
*/
func IsValidCategory(category string) bool {
	if category == CategoryUncategorized {
		return true
	}
	for _, c := range Categories {
		if c == category {
			return true
		}
	}
	return false
}
