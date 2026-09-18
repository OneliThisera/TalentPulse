package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"talentpulse/internal/ai"
	"talentpulse/internal/middleware"
	"talentpulse/internal/models"
	"talentpulse/internal/repository"
)

type Handler struct {
	store *repository.Store
	ai    *ai.AIService
}

func NewHandler(store *repository.Store, aiService *ai.AIService) *Handler {
	return &Handler{
		store: store,
		ai:    aiService,
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// Register creates a new user
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.FullName == "" {
		writeError(w, http.StatusBadRequest, "email, password, and full name are required")
		return
	}

	if req.Role != models.RoleCandidate && req.Role != models.RoleEmployer {
		req.Role = models.RoleCandidate
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hashed),
		FullName:     req.FullName,
		Role:         req.Role,
		CompanyName:  req.CompanyName,
		Title:        req.Title,
		Skills:       req.Skills,
		ResumeText:   req.ResumeText,
	}

	if err := h.store.CreateUser(&user); err != nil {
		writeError(w, http.StatusConflict, "user with this email already exists")
		return
	}

	token, err := middleware.GenerateToken(&user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate auth token")
		return
	}

	writeJSON(w, http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  user,
	})
}

// Login authenticates a user
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.store.GetUserByEmail(req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := middleware.GenerateToken(user)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

// GetMe returns the authenticated user profile
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.store.GetUserByID(claims.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// UpdateMe updates candidate skills, title, bio, and resume
func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Title      string `json:"title"`
		Bio        string `json:"bio"`
		Skills     string `json:"skills"`
		ResumeText string `json:"resume_text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.store.UpdateUserProfile(claims.UserID, req.Title, req.Bio, req.Skills, req.ResumeText); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	updated, _ := h.store.GetUserByID(claims.UserID)
	writeJSON(w, http.StatusOK, updated)
}

// ListJobs returns jobs with optional search and filters
func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("q")
	workplace := q.Get("workplace")
	exp := q.Get("experience")

	jobs, err := h.store.ListJobs(search, workplace, exp)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch jobs: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, jobs)
}

// GetJob returns a specific job
func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	job, err := h.store.GetJobByID(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}
	writeJSON(w, http.StatusOK, job)
}

// CreateJob allows employers to post a new job
func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil || claims.Role != models.RoleEmployer {
		writeError(w, http.StatusForbidden, "only employers can post jobs")
		return
	}

	var job models.JobPosting
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	job.EmployerID = claims.UserID
	job.EmployerName = claims.FullName
	if job.CompanyName == "" {
		u, _ := h.store.GetUserByID(claims.UserID)
		if u != nil && u.CompanyName != "" {
			job.CompanyName = u.CompanyName
		} else {
			job.CompanyName = "TalentCorp"
		}
	}
	if job.Currency == "" {
		job.Currency = "USD"
	}

	if err := h.store.CreateJob(&job); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create job posting")
		return
	}

	writeJSON(w, http.StatusCreated, job)
}

// ApplyJob submits a candidate application with AI fit scoring
func (h *Handler) ApplyJob(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil || claims.Role != models.RoleCandidate {
		writeError(w, http.StatusForbidden, "only candidates can apply to jobs")
		return
	}

	idStr := chi.URLParam(r, "id")
	jobID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	job, err := h.store.GetJobByID(jobID)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	var req struct {
		ResumeSummary string `json:"resume_summary"`
		CoverLetter   string `json:"cover_letter"`
		CandidateSkills string `json:"candidate_skills"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// If resume summary not provided, use user's stored resume
	if req.ResumeSummary == "" || req.CandidateSkills == "" {
		cand, _ := h.store.GetUserByID(claims.UserID)
		if cand != nil {
			if req.ResumeSummary == "" {
				req.ResumeSummary = cand.ResumeText
			}
			if req.CandidateSkills == "" {
				req.CandidateSkills = cand.Skills
			}
		}
	}

	// Compute AI match score and insights
	scoreRes, err := h.ai.ScoreCandidate(r.Context(), job, req.ResumeSummary, req.CandidateSkills)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ai scoring error")
		return
	}

	strengthsJSON, _ := json.Marshal(scoreRes.Strengths)
	gapsJSON, _ := json.Marshal(scoreRes.MissingSkills)

	app := models.Application{
		JobID:            jobID,
		CandidateID:      claims.UserID,
		ResumeSummary:    req.ResumeSummary,
		CoverLetter:      req.CoverLetter,
		AIMatchScore:     scoreRes.MatchScore,
		AIStrengths:      string(strengthsJSON),
		AIGapAnalysis:    string(gapsJSON),
		AIRecommendation: scoreRes.Recommendation,
		Notes:            scoreRes.SummaryAnalysis,
	}

	if err := h.store.CreateApplication(&app); err != nil {
		writeError(w, http.StatusConflict, "you have already applied to this job")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"application": app,
		"ai_analysis": scoreRes,
	})
}

// GetJobApplicants returns all applicants for a job (employer only)
func (h *Handler) GetJobApplicants(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil || claims.Role != models.RoleEmployer {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	idStr := chi.URLParam(r, "id")
	jobID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	apps, err := h.store.GetApplicationsByJob(jobID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get applications")
		return
	}

	writeJSON(w, http.StatusOK, apps)
}

// GetMyApplications returns candidate's submitted applications
func (h *Handler) GetMyApplications(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	apps, err := h.store.GetApplicationsByCandidate(claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch applications")
		return
	}

	writeJSON(w, http.StatusOK, apps)
}

// UpdateApplicationStatus changes candidate status in pipeline (Screening, Interview, Offer)
func (h *Handler) UpdateApplicationStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetUserFromContext(r.Context())
	if claims == nil || claims.Role != models.RoleEmployer {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	idStr := chi.URLParam(r, "appId")
	appID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid application id")
		return
	}

	var req struct {
		Status models.ApplicationStatus `json:"status"`
		Notes  string                   `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if err := h.store.UpdateApplicationStatus(appID, req.Status, req.Notes); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update application")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "status updated successfully"})
}

// AIGenerateJD generates a job description with AI
func (h *Handler) AIGenerateJD(w http.ResponseWriter, r *http.Request) {
	var req models.GenerateJDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	res, err := h.ai.GenerateJobDescription(r.Context(), &req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ai generation failed")
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// AIMatchPreview computes match preview before applying
func (h *Handler) AIMatchPreview(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	jobID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	job, err := h.store.GetJobByID(jobID)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	var req struct {
		ResumeText      string `json:"resume_text"`
		CandidateSkills string `json:"candidate_skills"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	scoreRes, err := h.ai.ScoreCandidate(r.Context(), job, req.ResumeText, req.CandidateSkills)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "scoring failed")
		return
	}
	writeJSON(w, http.StatusOK, scoreRes)
}

// AIInterviewQuestions generates interview questions for employer
func (h *Handler) AIInterviewQuestions(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	jobID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	job, err := h.store.GetJobByID(jobID)
	if err != nil {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	var req struct {
		ResumeText string `json:"resume_text"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	questions, err := h.ai.GenerateScreeningQuestions(r.Context(), job, req.ResumeText)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate questions")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"job_title": job.Title,
		"questions": questions,
	})
}
