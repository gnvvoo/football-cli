package cmd

import (
	"football-cli/internal/api"
	"football-cli/internal/output"
	"os"

	"github.com/spf13/cobra"
)

// predict 커맨드 flag 변수
var (
	predictHome    string
	predictAway    string
	predictExplain bool
)

var predictCmd = &cobra.Command{
	Use:   "predict",
	Short: "경기 결과 예측",
	Long:  "두 팀의 순위, 맞대결 데이터를 기반으로 경기 결과 예측. --home, --away 필수.",
	RunE:  runPredict,
}

func init() {
	rootCmd.AddCommand(predictCmd)

	predictCmd.Flags().StringVar(&predictHome, "home", "", "홈팀 이름 (부분 검색 가능) [필수]")
	predictCmd.Flags().StringVar(&predictAway, "away", "", "어웨이팀 이름 (부분 검색 가능) [필수]")
	predictCmd.Flags().BoolVar(&predictExplain, "explain", false, "예측 근거 출력")

	predictCmd.MarkFlagRequired("home")
	predictCmd.MarkFlagRequired("away")
}

func runPredict(cmd *cobra.Command, args []string) error {
	client, err := api.NewClient(Timeout)
	if err != nil {
		PrintError("AUTH_FAILURE", err.Error(), nil)
		os.Exit(ExitAuthFailure)
	}

	resp, err := client.Predict(predictHome, predictAway, predictExplain)
	if err != nil {
		switch err.Error() {
		case "NO_DATA":
			PrintError("NO_DATA", "팀을 찾을 수 없습니다. 팀 이름을 확인해주세요.", nil)
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

	output.PrintPredict(resp.Match.HomeTeam, resp.Match.AwayTeam,
		resp.Prediction.Winner, resp.Prediction.Confidence,
		resp.Prediction.Scores.Home, resp.Prediction.Scores.Away,
		resp.Basis.HomeForm, resp.Basis.AwayForm, resp.Basis.H2HSummary,
		resp.Reasoning,
	)

	return nil
}