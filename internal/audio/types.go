package audio

// PlayerStatus represents the current state of the audio player
type PlayerStatus struct {
	IsPlaying bool    `json:"is_playing"`
	IsPaused  bool    `json:"is_paused"`
	Position  string  `json:"position"` // Current playback position (e.g. "2:34")
	Duration  string  `json:"duration"` // Total track duration (e.g. "4:12")
	Volume    float64 `json:"volume"`   // Volume level 0.0 - 1.0
}