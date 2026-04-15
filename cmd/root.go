package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// 공통 flag 값을 담는 전역 변수
// var 블록으로 여러 전역 변수를 한 번에 선언하는 문법
// -> cmd 패키지 전체에서 접근 가능하다.
var (
	JSONOutput bool
	NoColor    bool
	Quiet      bool
	Timeout    int
)

// exit code 상수
// 상수 선언
const (
	ExitSuccess      = 0
	ExitGeneralError = 1
	ExitInvalidArgs  = 2
	ExitNoData       = 3
	ExitAPIFailure   = 4
	ExitAuthFailure  = 5
)

// cobra의 최상위 커맨드 객체
// &는 포인터 의미 -> cobra가 이 구조체를 직접 수정할 수 있도록 주소를 넘기는 것
var rootCmd = &cobra.Command{
	Use:   "football-cli",                                                                                // 명령어 이름
	Short: "AI-agent-friendly football data CLI",                                                         // --help에서 한 줄 설명
	Long:  "Fetch match schedules, standings, player stats, and predictions for top 5 European leagues.", // 상세 설명
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// main.go에서 호출한 함수 : cmd.Excute()
// rootCmd.Excute() 가  실제로 CLI 실행
// := 는 타입을 자동 추론해서 변수 선언
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(ExitGeneralError)
	}
}

// init() : go가 패키지 로딩 시 자동으로 호출하는 특수 함수.
// PersistentFlags() : 이 커맨드와 모든 하위 커맨드에 공통 적용되는 flag 등록
// &JSONOutput 등 : flag값이 저장될 변수의 주소. cobra가 직접 이 주소에 값을 써준다.
func init() {
	rootCmd.PersistentFlags().BoolVar(&JSONOutput, "json", false, "Output result as JSON")
	rootCmd.PersistentFlags().BoolVar(&NoColor, "no-color", false, "Disable ANSI color codes")
	rootCmd.PersistentFlags().BoolVar(&Quiet, "quiet", false, "Suppress stderr logs")
	rootCmd.PersistentFlags().IntVar(&Timeout, "timeout", 5000, "Request timeout in milliseconds")
}

// 에러 출력 헬퍼 — 항상 stderr로 출력
func PrintError(code string, message string, suggestions []string) {
	type ErrorBody struct {
		Code        string   `json:"code"`
		Message     string   `json:"message"`
		Suggestions []string `json:"suggestions,omitempty"`
	}
	type ErrorResponse struct {
		Error ErrorBody `json:"error"`
	}

	resp := ErrorResponse{
		Error: ErrorBody{
			Code:        code,
			Message:     message,
			Suggestions: suggestions,
		},
	}

	b, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Fprintln(os.Stderr, string(b))
}
