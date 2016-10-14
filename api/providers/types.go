package providers

//Provider represents a metric provider
type Provider interface {
	Connect() error
	CurrentCount() (int64, error)
}
