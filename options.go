package env

// Default option values
const (
	DefaultSeparator      = "__"
	DefaultSliceSeparator = ","
)

// Default option values
var (
	DefaultMapper = IdentityMapper
)

// Options is the set of decode options
type Options struct {
	Prefix         string              // String prefix that is stripped from each entry
	Separator      string              // String value for nested values
	SliceSeparator string              // Separator for slices encoded as strings
	Mapper         func(string) string // Mapping function for struct key names
}

func (o *Options) setDefaults() {
	if o.Separator == "" {
		o.Separator = DefaultSeparator
	}

	if o.SliceSeparator == "" {
		o.SliceSeparator = DefaultSliceSeparator
	}

	if o.Mapper == nil {
		o.Mapper = DefaultMapper
	}
}
