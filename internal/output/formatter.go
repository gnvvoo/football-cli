package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// PrintJSON : 데이터를 JSON 형식으로 stdout에 출력
func PrintJSON(data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// FormatMatchDate : ISO8601 날짜 문자열을 읽기 좋은 형식으로 변환
// "2026-04-13T15:00:00+00:00" → "2026-04-13 15:04"
func FormatMatchDate(iso string) string {
	t, err := time.Parse(time.RFC3339, iso)
	if err != nil {
		return iso // 파싱 실패시 원본 반환
	}
	
	// 한국 시간대 로드 (UTC+9)
	kst, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		// 시간대 로드 실패시 수동으로 UTC+9 적용
		kst = time.FixedZone("KST", 9*60*60)
	}

	return t.In(kst).Format("2006-01-02 15:04 (KST)")
}

// FormatScore : 점수를 읽기 좋은 형식으로 변환
// nil(경기 전) → "-"
// 1, 0 → "1 - 0"
func FormatScore(home, away *int) string {
	if home == nil || away == nil {
		return "-"
	}
	return fmt.Sprintf("%d - %d", *home, *away)
}

// FormatStatus : API 상태 코드를 한글로 변환
func FormatStatus(status string) string {
	switch status {
	case "NS":
		return "예정"
	case "1H":
		return "전반"
	case "HT":
		return "하프타임"
	case "2H":
		return "후반"
	case "ET":
		return "연장"
	case "P":
		return "승부차기"
	case "FT":
		return "종료"
	case "AET":
		return "연장종료"
	case "PEN":
		return "승부차기종료"
	case "PST":
		return "연기"
	default:
		return status
	}
}

// PrintMatchesTable : 경기 목록을 텍스트 테이블로 출력
func PrintMatchesTable(matches []MatchRow) {
	if len(matches) == 0 {
		fmt.Fprintln(os.Stderr, "경기 데이터가 없습니다.")
		return
	}

	// 헤더
	fmt.Printf("%-20s %-25s %-25s %-8s %-6s\n",
		"날짜", "홈팀", "어웨이팀", "상태", "점수")
	fmt.Println(strings.Repeat("-", 90))

	// 각 경기 출력
	for _, m := range matches {
		fmt.Printf("%-20s %-25s %-25s %-8s %-6s\n",
			m.Date,
			m.HomeTeam,
			m.AwayTeam,
			m.Status,
			m.Score,
		)
	}
}

// PrintStandingsTable : 순위 목록을 텍스트 테이블로 출력
func PrintStandingsTable(league, season string, rows []StandingRow) {
	if len(rows) == 0 {
		fmt.Fprintln(os.Stderr, "순위 데이터가 없습니다.")
		return
	}

	fmt.Printf("\n%s %s 시즌 순위\n", league, season)
	fmt.Println(strings.Repeat("-", 90))
	fmt.Printf("%-4s %-28s %-4s %-4s %-4s %-4s %-4s %-4s %-4s %-4s %-6s\n",
		"순위", "팀", "경기", "승", "무", "패", "득", "실", "득실", "승점", "최근폼")
	fmt.Println(strings.Repeat("-", 90))

	for _, r := range rows {
		fmt.Printf("%-4d %-28s %-4d %-4d %-4d %-4d %-4d %-4d %-4d %-4d %-6s\n",
			r.Rank, r.Team, r.Played, r.Won, r.Drawn, r.Lost,
			r.GF, r.GA, r.GD, r.Points, r.Form,
		)
	}
}

// PrintTeamInfo : 팀 정보 출력
func PrintTeamInfo(name string, founded int, venue string, leagues []string) {
	fmt.Println()
	fmt.Printf("팀명    : %s\n", name)
	fmt.Printf("창단    : %d년\n", founded)
	fmt.Printf("경기장  : %s\n", venue)
	fmt.Printf("리그    : %s\n", strings.Join(leagues, ", "))
}

// PrintPlayerStats : 선수 정보 출력
func PrintPlayerStats(name, position, dob, nationality, team string) {
	fmt.Println()
	fmt.Printf("이름    : %s\n", name)
	fmt.Printf("포지션  : %s\n", position)
	fmt.Printf("생년월일: %s\n", dob)
	fmt.Printf("국적    : %s\n", nationality)
	fmt.Printf("소속팀  : %s\n", team)
}

// MatchRow : 테이블 출력용 경기 데이터
type MatchRow struct {
	Date     string
	HomeTeam string
	AwayTeam string
	Status   string
	Score    string
}

// StandingRow : 테이블 출력용 순위 데이터
type StandingRow struct {
	Rank   int
	Team   string
	Played int
	Won    int
	Drawn  int
	Lost   int
	GF     int
	GA     int
	GD     int
	Points int
	Form   string
}