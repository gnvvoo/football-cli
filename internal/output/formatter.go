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

// MatchRow : 테이블 출력용 경기 데이터
type MatchRow struct {
	Date     string
	HomeTeam string
	AwayTeam string
	Status   string
	Score    string
}
