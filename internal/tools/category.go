package tools

// CategoryOrder is the order of groups in the audit, the list, and the README.
//
// The order is explicit. Sorting category names alphabetically would move a
// new category such as "Build Tools" between "All The Things" and
// "Cloud & Deployment" without anyone choosing that place. Add a category
// here in the same change that adds its catalog function. Registry
// concatenates those functions in this order.
var CategoryOrder = []string{
	"All The Things",
	"Cloud & Deployment",
	"Framework",
	"Package Manager",
	"Runtimes & SDKs",
	"Security & Observability",
}
