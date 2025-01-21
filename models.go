package main

type ESPNResponse struct {
	Events []Event `json:"events"`
}

type Event struct {
	ID           string        `json:"id"`
	Date         string        `json:"date"`
	Name         string        `json:"name"`
	ShortName    string        `json:"shortName"`
	Competitions []Competition `json:"competitions"`
	Status       Status        `json:"status"`
}

type Competition struct {
	ID          string       `json:"id"`
	Date        string       `json:"date"`
	Attendance  int          `json:"attendance"`
	Competitors []Competitor `json:"competitors"`
	Venue       Venue        `json:"venue"`
	Status      Status       `json:"status"`
}

type Competitor struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Order      int         `json:"order"`
	HomeAway   string      `json:"homeAway"`
	Winner     bool        `json:"winner"`
	Team       Team        `json:"team"`
	Score      string      `json:"score"`
	Linescores []Linescore `json:"linescores"`
}

type Linescore struct {
	Value float64 `json:"value"`
}

type Team struct {
	ID           string `json:"id"`
	Location     string `json:"location"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	DisplayName  string `json:"displayName"`
}

type Status struct {
	Clock        float64    `json:"clock"`
	DisplayClock string     `json:"displayClock"`
	Period       int        `json:"period"`
	Type         StatusType `json:"type"`
}

type StatusType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	Completed   bool   `json:"completed"`
	Description string `json:"description"`
}

type Venue struct {
	FullName string  `json:"fullName"`
	Address  Address `json:"address"`
}

type Address struct {
	City  string `json:"city"`
	State string `json:"state"`
}

type ScoreData struct {
	GameID        string  `json:"game_id"`
	Score         string  `json:"score"`
	QuarterScores []int64 `json:"quarter_scores"`
}
