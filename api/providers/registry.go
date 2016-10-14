package providers

//Creator construction function
type Creator func() Provider

//Providers all registered providers
var Providers = map[string]Creator{}

//AddProvider adds the provider type to map
func AddProvider(name string, creator Creator) {
	Providers[name] = creator
}
