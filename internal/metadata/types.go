package metadata

// Metadata holds track metadata information
type Metadata struct {
	Title   string `json:"title,omitempty"`
	Artist  string `json:"artist,omitempty"`
	Album   string `json:"album,omitempty"`
	Year    string `json:"year,omitempty"`
	Label   string `json:"label,omitempty"`   // Record label (TPUB)
	Rating  int    `json:"rating,omitempty"`  // 0-5 rating
	Genre   string `json:"genre,omitempty"`
	Comment string `json:"comment,omitempty"`
}

// IsEmpty returns true if no metadata is set
func (m *Metadata) IsEmpty() bool {
	return m.Title == "" &&
		m.Artist == "" &&
		m.Album == "" &&
		m.Year == "" &&
		m.Label == "" &&
		m.Rating == 0 &&
		m.Genre == "" &&
		m.Comment == ""
}
