package schema

// Manifest : CLI 전체 구조를 정의하는 스키마
// 에이전트가 --manifest 호출 시 이 구조체를 JSON으로 반환
type Manifest struct {
	Name        string                `json:"name"`
	Version     string                `json:"version"`
	Description string                `json:"description"`
	GlobalFlags map[string]FlagInfo   `json:"global_flags"`
	ExitCodes   map[string]string     `json:"exit_codes"`
	ErrorShape  ErrorShape            `json:"error_shape"`
	Leagues     map[string]LeagueInfo `json:"leagues"`
	Commands    map[string]Command    `json:"commands"`
}

// FlagInfo : flag 정보
type FlagInfo struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Default     any      `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// ErrorShape : 에러 응답 구조 설명
type ErrorShape struct {
	Error ErrorInfo `json:"error"`
}

// ErrorInfo : 에러 필드 설명
type ErrorInfo struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Suggestions string `json:"suggestions"`
}

// LeagueInfo : 리그 정보
type LeagueInfo struct {
	FullName string `json:"full_name"`
	Country  string `json:"country"`
	APIID    int    `json:"api_id"`
}

// Command : 각 커맨드 정보
type Command struct {
	Description string                 `json:"description"`
	Flags       map[string]CommandFlag `json:"flags"`
	Output      map[string]any         `json:"output"`
}

// CommandFlag : 커맨드 flag 정보
type CommandFlag struct {
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Format      string   `json:"format,omitempty"`
	Default     any      `json:"default,omitempty"`
}

// GetManifest : manifest 데이터 반환
func GetManifest() Manifest {
	leagues := []string{"EPL", "LaLiga", "Bundesliga", "SerieA", "Ligue1"}

	return Manifest{
		Name:        "football-cli",
		Version:     "1.0.0",
		Description: "AI-agent-friendly football data CLI for top 5 European leagues",
		GlobalFlags: map[string]FlagInfo{
			"--json":     {Type: "boolean", Description: "Output result as JSON"},
			"--no-color": {Type: "boolean", Description: "Disable ANSI color codes"},
			"--quiet":    {Type: "boolean", Description: "Suppress stderr logs"},
			"--timeout":  {Type: "integer", Description: "Request timeout in milliseconds", Default: 5000},
		},
		ExitCodes: map[string]string{
			"0": "Success",
			"1": "General error",
			"2": "Invalid arguments",
			"3": "No data found",
			"4": "External API failure",
			"5": "Authentication failure",
		},
		ErrorShape: ErrorShape{
			Error: ErrorInfo{
				Code:        "string — machine-readable error identifier e.g. TEAM_NOT_FOUND",
				Message:     "string — human-readable description",
				Suggestions: "array<string> — optional, present when alternatives exist",
			},
		},
		Leagues: map[string]LeagueInfo{
			"EPL":        {FullName: "Premier League", Country: "England", APIID: 2021},
			"LaLiga":     {FullName: "La Liga", Country: "Spain", APIID: 2014},
			"Bundesliga": {FullName: "Bundesliga", Country: "Germany", APIID: 2002},
			"SerieA":     {FullName: "Serie A", Country: "Italy", APIID: 2019},
			"Ligue1":     {FullName: "Ligue 1", Country: "France", APIID: 2015},
		},
		Commands: map[string]Command{
			"matches": {
				Description: "경기 일정 및 결과 조회",
				Flags: map[string]CommandFlag{
					"--league": {Type: "string", Required: true, Enum: leagues},
					"--date":   {Type: "string", Required: false, Format: "YYYY-MM-DD", Default: "today"},
					"--from":   {Type: "string", Required: false, Format: "YYYY-MM-DD"},
					"--to":     {Type: "string", Required: false, Format: "YYYY-MM-DD"},
					"--team":   {Type: "string", Required: false, Description: "팀 이름 (부분 검색 가능)"},
					"--status": {Type: "string", Required: false, Enum: []string{"live", "upcoming", "finished"}, Default: "all"},
				},
				Output: map[string]any{
					"matches":        "array<match>",
					"data_freshness": "string (ISO8601)",
					"match": map[string]string{
						"id":        "string",
						"date":      "string (ISO8601)",
						"status":    "string — live | upcoming | finished | postponed",
						"home_team": "string",
						"away_team": "string",
						"score":     "{ home: number|null, away: number|null }",
						"league":    "string",
						"venue":     "string",
					},
				},
			},
			"standings": {
				Description: "리그 순위 조회",
				Flags: map[string]CommandFlag{
					"--league": {Type: "string", Required: true, Enum: leagues},
					"--season": {Type: "integer", Required: false, Description: "시즌 시작 연도 e.g. 2024", Default: "current"},
				},
				Output: map[string]any{
					"league":         "string",
					"season":         "integer",
					"standings":      "array<team_standing>",
					"data_freshness": "string (ISO8601)",
					"team_standing": map[string]string{
						"rank":   "integer",
						"team":   "string",
						"played": "integer",
						"won":    "integer",
						"drawn":  "integer",
						"lost":   "integer",
						"gf":     "integer",
						"ga":     "integer",
						"gd":     "integer",
						"points": "integer",
						"form":   "string — e.g. WWDLW",
					},
				},
			},
			"player-stats": {
				Description: "선수 정보 조회",
				Flags: map[string]CommandFlag{
					"--player": {Type: "string", Required: true, Description: "선수 이름 (부분 검색 가능)"},
					"--season": {Type: "integer", Required: false, Default: "current"},
				},
				Output: map[string]any{
					"player": map[string]string{
						"id":            "string",
						"name":          "string",
						"position":      "string",
						"date_of_birth": "string",
						"nationality":   "string",
						"team":          "string",
					},
					"data_freshness": "string (ISO8601)",
				},
			},
			"team-info": {
				Description: "팀 정보 조회",
				Flags: map[string]CommandFlag{
					"--team": {Type: "string", Required: true, Description: "팀 이름 (부분 검색 가능)"},
				},
				Output: map[string]any{
					"team": map[string]string{
						"id":      "string",
						"name":    "string",
						"founded": "integer",
						"venue":   "string",
						"leagues": "array<string>",
					},
					"data_freshness": "string (ISO8601)",
				},
			},
			"predict": {
				Description: "경기 결과 예측",
				Flags: map[string]CommandFlag{
					"--home":    {Type: "string", Required: true},
					"--away":    {Type: "string", Required: true},
					"--explain": {Type: "boolean", Required: false, Default: false, Description: "예측 근거 출력"},
				},
				Output: map[string]any{
					"match": map[string]string{
						"home_team": "string",
						"away_team": "string",
					},
					"prediction": map[string]any{
						"winner":     "string — home | away | draw",
						"confidence": "number (0.0–1.0)",
						"scores": map[string]string{
							"home": "number",
							"away": "number",
						},
					},
					"basis": map[string]string{
						"home_form":   "string",
						"away_form":   "string",
						"h2h_summary": "string",
					},
					"reasoning":      "array<string> — only present when --explain is set",
					"data_freshness": "string (ISO8601)",
				},
			},
		},
	}
}

// GetCommandSchema : 특정 커맨드의 스키마만 반환
func GetCommandSchema(commandName string) (Command, bool) {
	manifest := GetManifest()
	cmd, ok := manifest.Commands[commandName]
	return cmd, ok
}
