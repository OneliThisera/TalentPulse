package models

type UserRole string

const (
	RoleCandidate UserRole = "candidate"
	RoleEmployer  UserRole = "employer"
	RoleAdmin     UserRole = "admin"
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         UserRole  `json:"role"`
	CompanyName  string    `json:"company_name,omitempty"`
	Title        string    `json:"title,omitempty"`
	Bio          string    `json:"bio,omitempty"`
	Skills       string    `json:"skills,omitempty"` // comma-separated or json string
	ResumeText   string    `json:"resume_text,omitempty"`
	CreatedAt    string    `json:"created_at"`
}

type JobPosting struct {
	ID              int64     `json:"id"`
	EmployerID      int64     `json:"employer_id"`
	EmployerName    string    `json:"employer_name"`
	CompanyName     string    `json:"company_name"`
	Title           string    `json:"title"`
	Department      string    `json:"department"`
	Location        string    `json:"location"`
	WorkplaceType   string    `json:"workplace_type"` // Remote, Hybrid, On-site
	EmploymentType  string    `json:"employment_type"` // Full-time, Contract, Part-time
	ExperienceLevel string    `json:"experience_level"` // Junior, Mid, Senior, Lead
	SalaryMin       int       `json:"salary_min"`
	SalaryMax       int       `json:"salary_max"`
	Currency        string    `json:"currency"`
	Description     string    `json:"description"`
	Requirements    string    `json:"requirements"`
	RequiredSkills  string    `json:"required_skills"`
	Status          string    `json:"status"` // open, closed, paused
	ApplicantCount  int       `json:"applicant_count"`
	CreatedAt       string    `json:"created_at"`
	UpdatedAt       string    `json:"updated_at"`
}

type ApplicationStatus string

const (
	AppStatusApplied     ApplicationStatus = "applied"
	AppStatusReviewing   ApplicationStatus = "reviewing"
	AppStatusShortlisted ApplicationStatus = "shortlisted"
	AppStatusInterviewed ApplicationStatus = "interviewed"
	AppStatusOffered     ApplicationStatus = "offered"
	AppStatusRejected    ApplicationStatus = "rejected"
)

type Application struct {
	ID               int64             `json:"id"`
	JobID            int64             `json:"job_id"`
	JobTitle         string            `json:"job_title,omitempty"`
	CompanyName      string            `json:"company_name,omitempty"`
	CandidateID      int64             `json:"candidate_id"`
	CandidateName    string            `json:"candidate_name,omitempty"`
	CandidateEmail   string            `json:"candidate_email,omitempty"`
	CandidateSkills  string            `json:"candidate_skills,omitempty"`
	ResumeSummary    string            `json:"resume_summary"`
	CoverLetter      string            `json:"cover_letter"`
	Status           ApplicationStatus `json:"status"`
	AIMatchScore     int               `json:"ai_match_score"` // 0-100
	AIStrengths      string            `json:"ai_strengths"`
	AIGapAnalysis    string            `json:"ai_gap_analysis"`
	AIRecommendation string            `json:"ai_recommendation"`
	Notes            string            `json:"notes,omitempty"`
	CreatedAt        string            `json:"created_at"`
	UpdatedAt        string            `json:"updated_at"`
}

type AIScoreResult struct {
	MatchScore      int      `json:"match_score"`
	Strengths       []string `json:"strengths"`
	MissingSkills   []string `json:"missing_skills"`
	Recommendation  string   `json:"recommendation"` // High, Medium, Low
	SummaryAnalysis string   `json:"summary_analysis"`
	InterviewQuestions []string `json:"interview_questions"`
}

type GenerateJDRequest struct {
	RoleTitle       string `json:"role_title"`
	Industry        string `json:"industry"`
	ExperienceLevel string `json:"experience_level"`
	KeySkills       string `json:"key_skills"`
	WorkplaceType   string `json:"workplace_type"`
}

type GenerateJDResponse struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	Requirements   string `json:"requirements"`
	RequiredSkills string `json:"required_skills"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	FullName    string   `json:"full_name"`
	Role        UserRole `json:"role"`
	CompanyName string   `json:"company_name,omitempty"`
	Title       string   `json:"title,omitempty"`
	Skills      string   `json:"skills,omitempty"`
	ResumeText  string   `json:"resume_text,omitempty"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
