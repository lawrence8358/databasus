package availability

type EngineAvailability struct {
	Engine            Engine `json:"engine"`
	IsAvailable       bool   `json:"isAvailable"`
	UnavailableReason string `json:"unavailableReason,omitempty"`
}

type FeatureAvailability struct {
	Feature           Feature `json:"feature"`
	IsAvailable       bool    `json:"isAvailable"`
	UnavailableReason string  `json:"unavailableReason,omitempty"`
}

type Availability struct {
	Engines  []EngineAvailability  `json:"engines"`
	Features []FeatureAvailability `json:"features"`
}
