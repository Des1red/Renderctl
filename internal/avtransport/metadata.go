package avtransport

// MetadataForVendor returns CurrentURIMetaData for a given vendor.
// Empty string means "no metadata".
func MetadataForVendor(vendor string, t Target) string {
	switch vendor {
	case "samsung":
		return ""
	case "lg":
		return lgMetadata(t)
	case "sony":
		return sonyMetadata(t)
	case "philips":
		return philipsMetadata(t)
	default:
		// generic
		return ""
	}
}
