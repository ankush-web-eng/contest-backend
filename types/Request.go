package types

type Request struct {
	UserId   string
	Code     string
	Language string
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignUpRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

type CreateContestRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	StartTime   string `json:"start_time" binding:"required"`
	EndTime     string `json:"end_time" binding:"required"`

	IsPublic    bool `json:"is_public" binding:"required"`
	MaxDuration int  `json:"max_duration"`

	CreatorID uint `json:"creator_id" binding:"required"`

	Status      string `json:"status" binding:"required"`
	RatingFloor int    `json:"rating_floor"`
	RatingCeil  int    `json:"rating_ceil"`

	IsRated       bool   `json:"is_rated" binding:"required"`
	RatingType    string `json:"rating_type" binding:"required"`
	RatingKFactor int    `json:"rating_k_factor" binding:"required"`
}

type UpdateContestRequest struct {
	Problems []struct {
		ContestID   uint   `json:"contest_id"`
		Title       string `json:"title"`
		Description string `json:"description"`

		TimeLimit   int    `json:"time_limit"`
		MemoryLimit int    `json:"memory_limit"`
		Difficulty  string `json:"difficulty"`
		Score       int    `json:"score"`
		Rating      int    `json:"rating"`

		SampleInput    string `json:"sample_input"`
		SampleOutput   string `json:"sample_output"`
		TestCasesCount int    `json:"test_cases_count"`

		TestCases []struct {
			ProblemID uint   `json:"problem_id"`
			Input     string `json:"input"`
			Output    string `json:"output"`
			IsHidden  bool   `json:"is_hidden"`
		} `json:"test_cases"`
	} `json:"problems"`
	ContestId uint `json:"contest_id"`
}

type SubmitCodeRequest struct {
	UserId    string `json:"user_id" binding:"required"`
	ContestID uint   `json:"contest_id" binding:"required"`
	ProblemID uint   `json:"problem_id" binding:"required"`
	Language  string `json:"language" binding:"required"`
	Code      string `json:"code" binding:"required"`
}
