package api

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheEntry는 캐시 파일에 저장되는 구조체
// 데이터와 함께 만료 시간을 저장
type CacheEntry struct {
	ExpiresAt time.Time       `json:"expires_at"`
	Data      json.RawMessage `json:"data"`
}

// cacheDir은 운영체제 임시 디렉토리 아래 football-cli-cache 폴더를 사용
// C:\Users\...\AppData\Local\Temp\football-cli-cache
func cacheDir() string {
	return filepath.Join(os.TempDir(), "football-cli-cache")
}

// cacheKey는 엔드포인트 문자열을 MD5 해시로 변환해서 파일명으로 사용
// "/fixtures?league=39&season=2024" 같은 긴 문자열을 짧은 해시로
func cacheKey(endpoint string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(endpoint)))
}

// SaveCache는 API 응답 데이터를 파일로 저장
// ttl은 캐시 유효 시간 (예: 5*time.Minute = 5분)
func SaveCache(endpoint string, data json.RawMessage, ttl time.Duration) error {
	// 캐시 디렉토리 생성 (없으면 생성, 있으면 무시)
	if err := os.MkdirAll(cacheDir(), 0755); err != nil {
		return err
	}

	entry := CacheEntry{
		ExpiresAt: time.Now().Add(ttl),
		Data:      data,
	}

	b, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	path := filepath.Join(cacheDir(), cacheKey(endpoint))
	return os.WriteFile(path, b, 0644)
}

// LoadCache는 캐시 파일을 읽어서 유효하면 데이터를 반환
// 캐시가 없거나 만료됐으면 nil 반환 → 호출하는 쪽에서 nil이면 API 호출
func LoadCache(endpoint string) json.RawMessage {
	path := filepath.Join(cacheDir(), cacheKey(endpoint))

	b, err := os.ReadFile(path)
	if err != nil {
		return nil // 캐시 파일 없음
	}

	var entry CacheEntry
	if err := json.Unmarshal(b, &entry); err != nil {
		return nil // 파싱 실패
	}

	// 만료 확인
	if time.Now().After(entry.ExpiresAt) {
		os.Remove(path) // 만료된 파일 삭제
		return nil
	}

	return entry.Data
}
