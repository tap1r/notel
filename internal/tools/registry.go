package tools

// Registry returns the developer tools notel can audit.
//
// The catalog is split by category so adding a tool is a short edit in one
// file. CategoryOrder decides the group order. Each catalog function keeps
// its own tools in name order. The zero values of Default and OptOut mean
// "telemetry is on until every variable matches".
//
// Commands are printed as hints. notel does not run them and does not treat
// them as proof that telemetry is off.
func Registry() []Tool {
	groups := [][]Tool{
		catalogAll(),
		catalogCloud(),
		catalogFramework(),
		catalogPackage(),
		catalogRuntime(),
		catalogSecurity(),
	}
	var all []Tool
	for _, group := range groups {
		all = append(all, group...)
	}
	return all
}
