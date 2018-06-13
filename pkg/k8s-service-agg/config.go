package k8s_service_agg

// Config for k8s service agg
type Config struct {
	Namespace             string
	PrefixesAnnotation    string
	ServicePortAnnotation string
	TimeoutAnnotation     string
	InCluster             bool
}
