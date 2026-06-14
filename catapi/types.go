package catapi

// --- output types (emitted to the user) ---

// CatImage is one cat image result from TheCatAPI.
type CatImage struct {
	ID     string `kit:"id" json:"id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// Breed is one cat breed from TheCatAPI.
type Breed struct {
	ID             string `kit:"id"             json:"id"`
	Name           string `json:"name"`
	Origin         string `json:"origin"`
	Temperament    string `json:"temperament"`
	LifeSpan       string `json:"life_span"`
	WeightMetric   string `json:"weight_metric"`
	Intelligence   int    `json:"intelligence"`
	AffectionLevel int    `json:"affection_level"`
	EnergyLevel    int    `json:"energy_level"`
	ChildFriendly  int    `json:"child_friendly"`
	WikipediaURL   string `json:"wikipedia_url"`
}

// --- wire types (unexported, only used for JSON decoding) ---

type wireImage struct {
	ID     string `json:"id"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type wireWeight struct {
	Imperial string `json:"imperial"`
	Metric   string `json:"metric"`
}

type wireBreed struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Origin         string     `json:"origin"`
	Temperament    string     `json:"temperament"`
	LifeSpan       string     `json:"life_span"`
	Weight         wireWeight `json:"weight"`
	Intelligence   int        `json:"intelligence"`
	AffectionLevel int        `json:"affection_level"`
	EnergyLevel    int        `json:"energy_level"`
	ChildFriendly  int        `json:"child_friendly"`
	WikipediaURL   string     `json:"wikipedia_url"`
}
