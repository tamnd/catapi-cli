package catapi

// --- output types (emitted to the user) ---

// Image is one cat image result from TheCatAPI.
type Image struct {
	ID     string `kit:"id" json:"id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Breed  string `json:"breed"`  // first breed name, or ""
	Origin string `json:"origin"` // first breed origin, or ""
}

// Breed is one cat breed from TheCatAPI.
type Breed struct {
	ID          string `kit:"id" json:"id"`
	Name        string `json:"name"`
	Origin      string `json:"origin"`
	Temperament string `json:"temperament"`
	LifeSpan    string `json:"life_span"`
	Description string `json:"description"`
}

// Category is one image category from TheCatAPI.
type Category struct {
	ID   int    `kit:"id" json:"id"`
	Name string `json:"name"`
}

// --- wire types (unexported, only used for JSON decoding) ---

type wireImage struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Breeds []struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		Temperament string `json:"temperament"`
		Origin      string `json:"origin"`
		LifeSpan    string `json:"life_span"`
	} `json:"breeds"`
}

type wireBreed struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Origin      string `json:"origin"`
	Temperament string `json:"temperament"`
	LifeSpan    string `json:"life_span"`
	Description string `json:"description"`
}

type wireCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
