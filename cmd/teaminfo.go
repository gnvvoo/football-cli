package cmd

import (
	"football-cli/internal/api"
	"football-cli/internal/output"
	"football-cli/internal/schema"
	"os"

	"github.com/spf13/cobra"
)

// team-info 커맨드 flag 변수
var teamInfoTeam string

var teamInfoCmd = &cobra.Command{
	Use:   "team-info",
	Short: "팀 정보 조회",
	Long:  "팀 이름으로 팀 정보 조회. --team 필수.",
	RunE:  runTeamInfo,
}

func init() {
	rootCmd.AddCommand(teamInfoCmd)

	teamInfoCmd.Flags().StringVar(&teamInfoTeam, "team", "", "팀 이름 (부분 검색 가능) [필수]")
	teamInfoCmd.MarkFlagRequired("team")
}

func runTeamInfo(cmd *cobra.Command, args []string) error {
	// --schema 플래그 시 스키마 출력 후 종료
	if SchemaFlag {
		s, _ := schema.GetCommandSchema("team-info")
		return output.PrintJSON(s)
	}

	client, err := api.NewClient(Timeout)
	if err != nil {
		PrintError("AUTH_FAILURE", err.Error(), nil)
		os.Exit(ExitAuthFailure)
	}

	resp, err := client.GetTeamInfo(teamInfoTeam)
	if err != nil {
		switch err.Error() {
		case "NO_DATA":
			PrintError("NO_DATA", "팀을 찾을 수 없습니다: "+teamInfoTeam, nil)
			os.Exit(ExitNoData)
		case "AUTH_FAILURE":
			PrintError("AUTH_FAILURE", "API 인증에 실패했습니다. API 키를 확인해주세요.", nil)
			os.Exit(ExitAuthFailure)
		default:
			PrintError("API_FAILURE", err.Error(), nil)
			os.Exit(ExitAPIFailure)
		}
	}

	if JSONOutput {
		return output.PrintJSON(resp)
	}

	output.PrintTeamInfo(resp.Team.Name, resp.Team.Founded, resp.Team.Venue, resp.Team.Leagues)

	return nil
}
