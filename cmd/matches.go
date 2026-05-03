package cmd

import (
	"football-cli/internal/api"
	"football-cli/internal/output"
	"football-cli/internal/schema"
	"os"

	"github.com/spf13/cobra"
)

// matches 커맨드 flag 변수
var (
	matchesLeague string
	matchesDate   string
	matchesFrom   string
	matchesTo     string
	matchesTeam   string
	matchesStatus string
)

var matchesCmd = &cobra.Command{
	Use:   "matches",
	Short: "경기 일정 및 결과 조회",
	Long:  "리그별 경기 일정과 결과 조회. --league 필수.",
	RunE:  runMatches,
}

func init() {
	// matches 커맨드를 루트 커맨드에 등록
	rootCmd.AddCommand(matchesCmd)

	// matches 전용 flag 등록
	matchesCmd.Flags().StringVar(&matchesLeague, "league", "", "리그 (EPL|LaLiga|Bundesliga|SerieA|Ligue1) [필수]")
	matchesCmd.Flags().StringVar(&matchesDate, "date", "", "날짜 (YYYY-MM-DD, 기본값: 오늘)")
	matchesCmd.Flags().StringVar(&matchesFrom, "from", "", "시작 날짜 (YYYY-MM-DD)")
	matchesCmd.Flags().StringVar(&matchesTo, "to", "", "종료 날짜 (YYYY-MM-DD)")
	matchesCmd.Flags().StringVar(&matchesTeam, "team", "", "팀 이름 (부분 검색 가능)")
	matchesCmd.Flags().StringVar(&matchesStatus, "status", "", "경기 상태 (live|upcoming|finished)")

	// --league 필수 flag로 지정
	matchesCmd.MarkFlagRequired("league")
}

func runMatches(cmd *cobra.Command, args []string) error {
	// --schema 플래그 시 스키마 출력 후 종료
	if SchemaFlag {
		s, _ := schema.GetCommandSchema("matches")
		return output.PrintJSON(s)
	}

	// 리그 약어 → API ID 변환
	leagueID, ok := api.LeagueIDs[matchesLeague]
	if !ok {
		PrintError(
			"INVALID_LEAGUE",
			"유효하지 않은 리그입니다: "+matchesLeague,
			[]string{"EPL", "LaLiga", "Bundesliga", "SerieA", "Ligue1"},
		)
		os.Exit(ExitInvalidArgs)
	}

	// API 클라이언트 생성
	client, err := api.NewClient(Timeout)
	if err != nil {
		PrintError("AUTH_FAILURE", err.Error(), nil)
		os.Exit(ExitAuthFailure)
	}

	// 경기 데이터 조회
	resp, err := client.GetMatches(leagueID, matchesDate, matchesTeam, matchesStatus)
	if err != nil {
		switch err.Error() {
		case "NO_DATA":
			PrintError("NO_DATA", "조회된 경기가 없습니다.", nil)
			os.Exit(ExitNoData)
		case "AUTH_FAILURE":
			PrintError("AUTH_FAILURE", "API 인증에 실패했습니다. API 키를 확인해주세요.", nil)
			os.Exit(ExitAuthFailure)
		default:
			PrintError("API_FAILURE", err.Error(), nil)
			os.Exit(ExitAPIFailure)
		}
	}

	// --json 플래그 시 JSON 출력
	if JSONOutput {
		return output.PrintJSON(resp)
	}

	// 텍스트 테이블 출력
	var rows []output.MatchRow
	for _, m := range resp.Matches {
		rows = append(rows, output.MatchRow{
			Date:     output.FormatMatchDate(m.Date),
			HomeTeam: m.HomeTeam,
			AwayTeam: m.AwayTeam,
			Status:   output.FormatStatus(m.Status),
			Score:    output.FormatScore(m.Score.Home, m.Score.Away),
		})
	}
	output.PrintMatchesTable(rows)

	return nil
}
