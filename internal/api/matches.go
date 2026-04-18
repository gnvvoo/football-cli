package api

import (
	"fmt"
	"time"
)

// API-Football의 경기 데이터 구조체야
type Match struct {
	Fixture Fixture `json:"fixture"`
	League  League  `json:"league"`
	Teams   Teams   `json:"teams"`
	Goals   Goals   `json:"goals"`
}

// 경기 기본 정보야
type Fixture struct {
	ID     int    `json:"id"`
	Date   string `json:"date"`   // ISO8601 형식 "2026-04-13T15:00:00+00:00"
	Status Status `json:"status"` // 경기 상태
	Venue  Venue  `json:"venue"`  // 경기장 정보
}

// 경기 진행 상태
type Status struct {
	Short string `json:"short"` // "NS"=예정, "1H"=전반, "HT"=하프타임, "2H"=후반, "FT"=종료, "PST"=연기
	Long  string `json:"long"`  // "Not Started", "First Half" 등 전체 이름
}

// 경기장 정보
type Venue struct {
	Name string `json:"name"` // 경기장 이름
	City string `json:"city"` // 도시
}

// 리그 정보야
type League struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// 홈/어웨이 팀 정보야
type Teams struct {
	Home Team `json:"home"`
	Away Team `json:"away"`
}

// 팀 기본 정보야
type Team struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// 득점 정보
// 포인터(*int)를 쓰는 이유 : 경기 전엔 점수가 null이라서
// int면 null을 0으로 오해 가능성 → *int면 null = nil로 구분 가능
type Goals struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

// --manifest 스키마의 matches output 구조
type MatchesResponse struct {
	Matches       []MatchOutput `json:"matches"`
	DataFreshness string        `json:"data_freshness"`
}

// CLI가 최종 출력하는 경기 데이터 구조야
// API 응답을 -> 지정한 스키마에 맞게 변환 후 출력
type MatchOutput struct {
	ID       int         `json:"id"`
	Date     string      `json:"date"`
	Status   string      `json:"status"`
	HomeTeam string      `json:"home_team"`
	AwayTeam string      `json:"away_team"`
	Score    ScoreOutput `json:"score"`
	League   string      `json:"league"`
	Venue    string      `json:"venue"`
}

// 점수 출력 구조야
type ScoreOutput struct {
	Home *int `json:"home"`
	Away *int `json:"away"`
}

// API-Football에서 경기 목록을 가져오는 함수야
// leagueID: 리그 ID (예: 39)
// date: "2026-04-13" 형식, 빈 문자열이면 오늘 날짜 사용
// team: 팀 이름 필터 (빈 문자열이면 전체)
// status: "live", "upcoming", "finished" (빈 문자열이면 전체)
func (c *Client) GetMatches(leagueID int, date, team, status string) (*MatchesResponse, error) {
	// 날짜가 없으면 오늘 날짜 사용
	if date == "" {
		date = time.Now().Format("2006-01-02") // Go의 날짜 포맷은 2006-01-02 기준이야
	}

	// 엔드포인트 조합
	endpoint := fmt.Sprintf("/fixtures?league=%d&date=%s&season=%d",
		leagueID,
		date,
		currentSeason(),
	)

	// API 호출
	var matches []Match
	if err := c.Get(endpoint, &matches); err != nil {
		return nil, err
	}

	// API 응답을 우리 출력 구조체로 변환
	output := &MatchesResponse{
		DataFreshness: time.Now().UTC().Format(time.RFC3339),
	}

	for _, m := range matches {
		// 상태 필터링
		if status != "" && !matchStatus(m.Fixture.Status.Short, status) {
			continue
		}

		// 팀 필터링 (부분 문자열 매칭)
		if team != "" && !containsTeam(m.Teams, team) {
			continue
		}

		output.Matches = append(output.Matches, MatchOutput{
			ID:       m.Fixture.ID,
			Date:     m.Fixture.Date,
			Status:   m.Fixture.Status.Short,
			HomeTeam: m.Teams.Home.Name,
			AwayTeam: m.Teams.Away.Name,
			Score: ScoreOutput{
				Home: m.Goals.Home,
				Away: m.Goals.Away,
			},
			League: m.League.Name,
			Venue:  m.Fixture.Venue.Name,
		})
	}

	return output, nil
}

// 현재 시즌 연도를 반환해
// 8월 이전이면 전년도 시즌 -> 보통 시즌 8월 시작
func currentSeason() int {
	now := time.Now()
	if now.Month() < time.August {
		return now.Year() - 1
	}
	return now.Year()
}

// API 상태 코드를 status 필터에 맞게 변환
func matchStatus(apiStatus, filter string) bool {
	switch filter {
	case "live":
		// 1H=전반, HT=하프타임, 2H=후반, ET=연장, P=승부차기
		return apiStatus == "1H" || apiStatus == "HT" ||
			apiStatus == "2H" || apiStatus == "ET" || apiStatus == "P"
	case "upcoming":
		return apiStatus == "NS" // NS = Not Started
	case "finished":
		return apiStatus == "FT" || apiStatus == "AET" || apiStatus == "PEN"
	}
	return true // 필터 없으면 전체 반환
}

// 홈 또는 어웨이 팀 이름에 검색어가 포함되는지 확인
// 대소문자 구분 없이 부분 매칭
func containsTeam(teams Teams, query string) bool {
	import_lower := func(s string) string {
		result := ""
		for _, c := range s {
			if c >= 'A' && c <= 'Z' {
				result += string(c + 32)
			} else {
				result += string(c)
			}
		}
		return result
	}
	q := import_lower(query)
	return contains(import_lower(teams.Home.Name), q) ||
		contains(import_lower(teams.Away.Name), q)
}

// 문자열 포함 여부 확인
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(substr) == 0 ||
		findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
