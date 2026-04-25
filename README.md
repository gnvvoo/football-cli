# football-cli

AI 에이전트 친화적인 유럽 5대 리그 축구 데이터 CLI
(플랜의 한계로 아직 개발중입니다!)

---

## 소개

`football-cli`는 EPL, La Liga, Bundesliga, Serie A, Ligue 1의 경기 일정, 순위, 팀/선수 정보, 경기 예측을 제공하는 CLI 도구입니다.

AI 에이전트가 `--manifest` 진입점을 통해 전체 구조를 파악하고 자율적으로 데이터를 조회할 수 있도록 설계되었습니다.

---

## 설치

### 요구사항

- Go 1.21 이상
- football-data.org API 키 ([무료 발급](https://www.football-data.org))

### 빌드

```bash
git clone https://github.com/your-username/football-cli.git
cd football-cli
go build -o football-cli
```

### 환경변수 설정

프로젝트 루트에 `.env` 파일 생성:

```
FOOTBALL_DATA_API_KEY=발급받은_API_키
```

---

## 사용법

### 기본 구조

```
football-cli
 ├ --manifest         AI 에이전트 진입점 — 전체 CLI 구조 반환
 ├ matches            경기 일정 및 결과 조회
 ├ standings          리그 순위 조회
 ├ team-info          팀 정보 조회
 ├ player-stats       선수 정보 조회
 └ predict            경기 결과 예측
```

### 공통 플래그

| 플래그       | 설명                   | 기본값 |
| ------------ | ---------------------- | ------ |
| `--json`     | JSON 형식으로 출력     | false  |
| `--no-color` | ANSI 색상 코드 제거    | false  |
| `--quiet`    | stderr 로그 억제       | false  |
| `--timeout`  | API 요청 제한시간 (ms) | 5000   |

### 지원 리그

| 약어         | 리그           | 국가     |
| ------------ | -------------- | -------- |
| `EPL`        | Premier League | 잉글랜드 |
| `LaLiga`     | La Liga        | 스페인   |
| `Bundesliga` | Bundesliga     | 독일     |
| `SerieA`     | Serie A        | 이탈리아 |
| `Ligue1`     | Ligue 1        | 프랑스   |

---

## 커맨드

### --manifest

AI 에이전트 진입점. CLI 전체 구조, 스키마, exit code를 JSON으로 반환합니다.

```bash
football-cli --manifest
```

---

### matches

리그별 경기 일정 및 결과를 조회합니다.

```bash
football-cli matches --league EPL
football-cli matches --league EPL --date 2026-04-18
football-cli matches --league EPL --team Arsenal
football-cli matches --league EPL --status upcoming
football-cli matches --league EPL --json
```

| 플래그     | 필수 | 설명                               |
| ---------- | ---- | ---------------------------------- |
| `--league` | ✅   | 리그 약어                          |
| `--date`   |      | 날짜 (YYYY-MM-DD, 기본값: 오늘)    |
| `--from`   |      | 시작 날짜 (YYYY-MM-DD)             |
| `--to`     |      | 종료 날짜 (YYYY-MM-DD)             |
| `--team`   |      | 팀 이름 (부분 검색 가능)           |
| `--status` |      | `live` \| `upcoming` \| `finished` |

**출력 예시**

```
날짜                   홈팀                       어웨이팀                    상태     점수
------------------------------------------------------------------------------------------
2026-04-18 20:30     Brentford FC              Fulham FC                 종료   0 - 0
2026-04-18 23:00     Newcastle United FC       AFC Bournemouth           종료   1 - 2
```

---

### standings

리그 순위를 조회합니다.

```bash
football-cli standings --league EPL
football-cli standings --league LaLiga --json
```

| 플래그     | 필수 | 설명                               |
| ---------- | ---- | ---------------------------------- |
| `--league` | ✅   | 리그 약어                          |
| `--season` |      | 시즌 연도 (예: 2024, 기본값: 현재) |

**출력 예시**

```
Premier League 2025 시즌 순위
------------------------------------------------------------------------------------------
순위  팀                            경기  승   무   패   득   실   득실  승점
------------------------------------------------------------------------------------------
1    Manchester City FC           33   21   7    5    66   29   37   70
2    Arsenal FC                   33   21   7    5    63   26   37   70
```

---

### team-info

팀 정보를 조회합니다.

```bash
football-cli team-info --team Arsenal
football-cli team-info --team "Manchester City"
```

| 플래그   | 필수 | 설명                     |
| -------- | ---- | ------------------------ |
| `--team` | ✅   | 팀 이름 (부분 검색 가능) |

**출력 예시**

```
팀명    : Arsenal FC
창단    : 1886년
경기장  : Emirates Stadium
리그    : Premier League, UEFA Champions League
```

---

### player-stats

선수 정보를 조회합니다.

```bash
football-cli player-stats --player Salah
football-cli player-stats --player "Erling Haaland"
```

| 플래그     | 필수 | 설명                       |
| ---------- | ---- | -------------------------- |
| `--player` | ✅   | 선수 이름 (부분 검색 가능) |

**출력 예시**

```
이름    : Mohamed Salah
포지션  : Right Winger
생년월일: 1992-06-15
국적    : Egypt
소속팀  : Liverpool FC
```

---

### predict

두 팀의 순위, 맞대결 데이터를 기반으로 경기 결과를 예측합니다.

```bash
football-cli predict --home Arsenal --away Chelsea
football-cli predict --home Arsenal --away Chelsea --explain
football-cli predict --home "Manchester City" --away Liverpool --json
```

| 플래그      | 필수 | 설명                           |
| ----------- | ---- | ------------------------------ |
| `--home`    | ✅   | 홈팀 이름 (부분 검색 가능)     |
| `--away`    | ✅   | 어웨이팀 이름 (부분 검색 가능) |
| `--explain` |      | 예측 근거 상세 출력            |

**출력 예시**

```
[ Arsenal FC vs Chelsea FC ]
------------------------------------------------------------
예측 결과  : Arsenal FC 승 (신뢰도 63%)
예상 스코어: 2 - 1

근거
  홈팀  : 2위 (승점 70)
  어웨이: 8위 (승점 48)
  H2H   : 최근 맞대결 1경기: 1승 0무 0패

상세 분석 (--explain)
------------------------------------------------------------
  - Arsenal FC 현재 2위 (승점 70, 경기당 평균 2.12점)
  - Chelsea FC 현재 8위 (승점 48, 경기당 평균 1.41점)
  - 최근 맞대결: 1승 0무 0패
  - 홈 어드밴티지 +10% 적용
  - 홈팀 승 확률 63% / 무승부 7% / 원정팀 승 확률 30%
```

---

## Exit Code

| 코드 | 의미          |
| ---- | ------------- |
| `0`  | 성공          |
| `1`  | 일반 오류     |
| `2`  | 잘못된 인자   |
| `3`  | 데이터 없음   |
| `4`  | 외부 API 실패 |
| `5`  | 인증 실패     |

---

## 에러 응답 구조

`--json` 모드에서 에러 발생 시 항상 아래 구조로 반환됩니다.

```json
{
  "error": {
    "code": "TEAM_NOT_FOUND",
    "message": "팀을 찾을 수 없습니다: Manchaster",
    "suggestions": ["Manchester United", "Manchester City"]
  }
}
```

---

## AI 에이전트 연동

에이전트는 `--manifest`를 진입점으로 CLI 구조를 파악한 후 자율적으로 커맨드를 실행합니다.

```bash
# 1. 진입점 탐색
football-cli --manifest

# 2. 데이터 조회 (항상 --json --no-color --quiet 플래그 사용 권장)
football-cli matches --league EPL --json --no-color --quiet

# 3. exit code로 성공/실패 분기
# 0 = 성공, 3 = 데이터 없음, 4 = API 실패
```

---

## 프로젝트 구조

```
football-cli/
├ main.go
├ cmd/
│  ├ root.go          공통 플래그, exit code 정의
│  ├ matches.go        matches 커맨드
│  ├ standings.go      standings 커맨드
│  ├ teaminfo.go       team-info 커맨드
│  ├ playerstat.go     player-stats 커맨드
│  └ predict.go        predict 커맨드
├ internal/
│  ├ api/
│  │  ├ client.go      football-data.org HTTP 클라이언트
│  │  ├ cache.go       파일 캐시 (TTL 5분)
│  │  ├ matches.go     경기 데이터 조회
│  │  ├ standings.go   순위 데이터 조회
│  │  ├ teaminfo.go    팀 정보 조회
│  │  ├ playerstats.go 선수 정보 조회
│  │  └ predict.go     경기 예측 로직
│  ├ output/
│  │  └ formatter.go   텍스트/JSON 출력 포맷터
│  └ schema/
│     └ manifest.go    --manifest 스키마 정의
└ .env                 API 키 (gitignore)
```

---

## 데이터 소스

- **[football-data.org](https://www.football-data.org)** — 경기 일정, 순위, 팀/선수 정보
- 무료 플랜 기준 현재 시즌 데이터 제공

---

## 한계

- 선수 시즌 스탯 미제공 (football-data.org 무료 플랜 한계)
- 팀 최근 폼 미제공 (무료 플랜 한계)
- predict는 순위/맞대결 기반 통계 예측으로 참고용
