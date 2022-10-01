package util

func ToStringPtr(s string) *string    { return &s }
func ToInt64Ptr(i int64) *int64       { return &i }
func ToBoolPtr(b bool) *bool          { return &b }
func ToFloat64Ptr(f float64) *float64 { return &f }
func ToFloat32Ptr(f float32) *float32 { return &f }
func GetInt64Val(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func GetBoolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

func GetStringVal(t *string) string {
	if t == nil {
		return ""
	}
	return *t
}
