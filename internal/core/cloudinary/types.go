package cloudinary

type File struct {
	URL      string
	PublicID string
	Width    int
	Height   int
	Format   string
	Bytes    int
	Type     string
}
