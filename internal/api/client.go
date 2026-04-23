package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// API 기본 URL
const (
	BaseURL = "https://api.football-data.org/v4"
)

// 리그 약어 → API-Football 내부 ID 매핑
// map[string]int : key가 string, value가 int인 해시맵
var LeagueIDs = map[string]int{
	"EPL":        PL,
	"LaLiga":     PD,
	"Bundesliga": BL1,
	"SerieA":     SA,
	"Ligue1":     FL1,
}

const (
	PL  = 2021 // Premier League
	PD  = 2014 // La Liga
	BL1 = 2002 // Bundesliga
	SA  = 2019 // Serie A
	FL1 = 2015 // Ligue 1
)

// API-Football HTTP 클라이언트 구조체
// 다른 언어의 클래스와 비슷
// 모든 API 요청은 이 구조체의 메서드로 호출
type Client struct {
	apiKey     string       // API 인증 키
	httpClient *http.Client // 타임아웃이 설정된 HTTP 클라이언트
}

// Go는 클래스 생성자가 없기에 New~ 함수를 만드는 것이 관례
// -> NewClient : Client 구조체를 생성하는 생성자 함수
// 반환값 : (*Client, error)
func NewClient(timeoutMs int) (*Client, error) {
	apiKey := os.Getenv("FOOTBALL_DATA_API_KEY") // 환경변수에서 API 키를 읽어서 Client를 초기화
	// API 호출 실패시
	if apiKey == "" {
		// (*Client, error)에 (nil, Error) 반환
		return nil, fmt.Errorf("FOOTBALL_DATA_API_KEY 환경변수가 설정되지 않았습니다")
	}

	// API 호출 성공시
	// (*Client, error)에 (&Client, nil) 반환
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutMs) * time.Millisecond,
		},
	}, nil
}

// API-Football의 공통 응답 구조체
// 모든 엔드포인트가 이 형태로 응답을 감싸서 반환
// “ : 구조체 태그 -> JSON 파싱할 때 어떤 키와 매핑할지 지정
// ex) Get        string          `json:"get"`  <- json의 "get"키를 Go의 Get 필드에 매핑
// type APIResponse struct {
// 	Get        string          `json:"get"`        // 요청한 엔드포인트
// 	Parameters map[string]any  `json:"parameters"` // 요청 파라미터
// 	Errors     any             `json:"errors"`     // API 레벨 에러
// 	Results    int             `json:"results"`    // 결과 개수
// 	Response   json.RawMessage `json:"response"`   // 실제 데이터 (엔드포인트마다 구조가 달라서 RawMessage로 받음)
// }

// football-data.org 공통 응답 처리
// 이 API는 엔드포인트마다 응답 구조가 달라서 RawMessage로 받아서 각자 파싱
type RawResponse = json.RawMessage

// Get은 football-data.org에 GET 요청을 보내는 공통 함수
// 응답의 response 필드를 result에 파싱
// (c *Client)는 메서드 리시버. 이 함수가 Client 구조체에 속한다는 의미.
// c.apikey 처럼 함수내에서 사용하기 위함
func (c *Client) Get(endpoint string, result any) error {
	// 캐시 확인 — 유효한 캐시가 있으면 API 요청 생략
	if cached := LoadCache(endpoint); cached != nil {
		return json.Unmarshal(cached, result)
	}

	url := BaseURL + endpoint

	// HTTP 요청 객체 생성
	// := : 선언과 동시에 값 할당
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("요청 생성 실패: %w", err) // %w : 에러를 래핑하는 포맷
	}

	// API-Football 인증 헤더 설정
	req.Header.Set("X-Auth-Token", c.apiKey)

	// 요청 실행
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("API 요청 실패: %w", err)
	}
	// defer : 현재 함수가 끝날 때 실행할 코드 예약
	defer resp.Body.Close() // 함수 종료 시 응답 바디 반드시 닫기 -> 메모리 누수 방지

	// 인증 실패 처리
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("AUTH_FAILURE")
	}

	// 기타 HTTP 에러 처리
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API 응답 오류: HTTP %d", resp.StatusCode)
	}	

	// 응답 바디 읽기
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("응답 읽기 실패: %w", err)
	}

	// 결과 파싱
	// json.Unmarshal : JSON을 Go 구조체로 변환
	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("JSON 파싱 실패: %w", err)
	}

	// API 응답을 캐시에 저장 — 5분 TTL
	b, _ := json.Marshal(result)
	SaveCache(endpoint, b, 5*time.Minute)

	return nil
}
