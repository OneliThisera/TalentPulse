package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"talentpulse/internal/models"
)

type Engine interface {
	ScoreCandidate(ctx context.Context, job *models.JobPosting, resumeText string, candidateSkills string) (*models.AIScoreResult, error)
	GenerateJobDescription(ctx context.Context, req *models.GenerateJDRequest) (*models.GenerateJDResponse, error)
	GenerateScreeningQuestions(ctx context.Context, job *models.JobPosting, resumeText string) ([]string, error)
}

type AIService struct {
	apiKey     string
	provider   string // "gemini", "openai", or "local"
	httpClient *http.Client
}

func NewAIService() *AIService {
	apiKey := os.Getenv("GEMINI_API_KEY")
	provider := "gemini"
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey != "" {
			provider = "openai"
		} else {
			provider = "local"
		}
	}
	return &AIService{
		apiKey:     apiKey,
		provider:   provider,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// ScoreCandidate analyzes a candidate's resume and skills against job requirements
func (s *AIService) ScoreCandidate(ctx context.Context, job *models.JobPosting, resumeText string, candidateSkills string) (*models.AIScoreResult, error) {
	if s.provider == "gemini" && s.apiKey != "" {
		res, err := s.geminiScoreCandidate(ctx, job, resumeText, candidateSkills)
		if err == nil {
			return res, nil
		}
		// Fallback to local heuristic engine if API fails or rate limited
	}

	return s.localScoreCandidate(job, resumeText, candidateSkills), nil
}

// GenerateJobDescription auto-generates a rich, structured job description
func (s *AIService) GenerateJobDescription(ctx context.Context, req *models.GenerateJDRequest) (*models.GenerateJDResponse, error) {
	if s.provider == "gemini" && s.apiKey != "" {
		res, err := s.geminiGenerateJD(ctx, req)
		if err == nil {
			return res, nil
		}
	}

	return s.localGenerateJD(req), nil
}

// GenerateScreeningQuestions suggests interview questions tailored to candidate and job
func (s *AIService) GenerateScreeningQuestions(ctx context.Context, job *models.JobPosting, resumeText string) ([]string, error) {
	if s.provider == "gemini" && s.apiKey != "" {
		res, err := s.geminiQuestions(ctx, job, resumeText)
		if err == nil {
			return res, nil
		}
	}

	return s.localQuestions(job, resumeText), nil
}

// ---- Local Heuristic & NLP Fallback Engine ----

func parseSkills(raw string) []string {
	splits := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == '\n' || r == '/'
	})
	var clean []string
	seen := make(map[string]bool)
	for _, s := range splits {
		trimmed := strings.ToLower(strings.TrimSpace(s))
		if len(trimmed) > 1 && !seen[trimmed] {
			seen[trimmed] = true
			clean = append(clean, trimmed)
		}
	}
	return clean
}

func (s *AIService) localScoreCandidate(job *models.JobPosting, resumeText string, candidateSkills string) *models.AIScoreResult {
	reqSkills := parseSkills(job.RequiredSkills)
	if len(reqSkills) == 0 {
		reqSkills = parseSkills(job.Requirements)
	}

	candSkills := parseSkills(candidateSkills)
	normResume := strings.ToLower(resumeText)

	matchedSkills := []string{}
	missingSkills := []string{}

	for _, skill := range reqSkills {
		hasSkill := false
		for _, cs := range candSkills {
			if strings.Contains(cs, skill) || strings.Contains(skill, cs) {
				hasSkill = true
				break
			}
		}
		if !hasSkill && strings.Contains(normResume, skill) {
			hasSkill = true
		}

		if hasSkill {
			matchedSkills = append(matchedSkills, strings.Title(skill))
		} else {
			missingSkills = append(missingSkills, strings.Title(skill))
		}
	}

	score := 45 // baseline
	total := len(reqSkills)
	if total > 0 {
		matchedRatio := float64(len(matchedSkills)) / float64(total)
		score = int(matchedRatio * 60) + 35
	} else {
		score = 75
	}

	// Adjust based on experience keyword density
	expLevel := strings.ToLower(job.ExperienceLevel)
	if strings.Contains(normResume, expLevel) || strings.Contains(normResume, "years") || strings.Contains(normResume, "lead") {
		score += 5
	}
	if score > 98 {
		score = 98
	}
	if score < 25 {
		score = 25
	}

	recommendation := "Strong Match"
	if score < 60 {
		recommendation = "Low Match"
	} else if score < 80 {
		recommendation = "Moderate Match"
	}

	analysis := fmt.Sprintf("Candidate matches %d out of %d targeted core competencies. Profile demonstrates solid alignment for %s.", len(matchedSkills), total, job.Title)
	if len(missingSkills) > 0 {
		analysis += fmt.Sprintf(" Recommend verifying proficiency in: %s during technical screening.", strings.Join(missingSkills, ", "))
	}

	questions := []string{
		fmt.Sprintf("Can you walk us through a recent project where you applied %s in production?", strings.Join(takeFirst(matchedSkills, 2), " and ")),
		fmt.Sprintf("How would you quickly ramp up on %s for high-scale enterprise applications?", strings.Join(takeFirst(missingSkills, 2), " and ")),
		fmt.Sprintf("Describe an architectural challenge you resolved for a role similar to %s.", job.Title),
	}

	return &models.AIScoreResult{
		MatchScore:         score,
		Strengths:          matchedSkills,
		MissingSkills:      missingSkills,
		Recommendation:     recommendation,
		SummaryAnalysis:    analysis,
		InterviewQuestions: questions,
	}
}

func (s *AIService) localGenerateJD(req *models.GenerateJDRequest) *models.GenerateJDResponse {
	title := req.RoleTitle
	if title == "" {
		title = "Software Engineer"
	}
	skills := req.KeySkills
	if skills == "" {
		skills = "Go, PostgreSQL, Docker, Microservices, REST APIs, Kubernetes"
	}
	level := req.ExperienceLevel
	if level == "" {
		level = "Mid-Senior"
	}
	workplace := req.WorkplaceType
	if workplace == "" {
		workplace = "Hybrid / Remote Friendly"
	}

	desc := fmt.Sprintf(`### Role Overview
We are looking for an exceptional %s to join our fast-paced %s team. In this role, you will be responsible for architecting, building, and deploying resilient systems that drive our core customer experiences.

### Key Responsibilities
- Architect, build, and maintain high-performance, fault-tolerant backend services and APIs.
- Collaborate closely with product managers, designers, and fellow engineers to deliver impactful end-to-end features.
- Write maintainable, well-tested code adhering to modern best practices, code reviews, and automated CI/CD pipelines.
- Diagnose and optimize system latency, resource utilization, and operational scalability.
- Champion engineering excellence, mentoring team members, and contributing to technical architectural blueprints.`, title, req.Industry)

	reqs := fmt.Sprintf(`### Qualifications & Requirements
- %s experience delivering software solutions in high-velocity environments.
- Proven hands-on track record working with: %s.
- Deep understanding of distributed system patterns, data integrity, and API design.
- Passion for automation, automated testing, and robust error handling.
- Excellent written and verbal communication skills; ability to articulate complex engineering tradeoffs.
- Workplace environment: %s.`, level, skills, workplace)

	return &models.GenerateJDResponse{
		Title:          title,
		Description:    strings.TrimSpace(desc),
		Requirements:   strings.TrimSpace(reqs),
		RequiredSkills: skills,
	}
}

func (s *AIService) localQuestions(job *models.JobPosting, resumeText string) []string {
	return []string{
		fmt.Sprintf("Based on the requirements for %s, what architectural design patterns do you gravitate toward?", job.Title),
		"How do you handle zero-downtime database migrations and graceful degradations in distributed services?",
		"Describe a situation where you identified and resolved a critical performance bottleneck in production.",
		"What is your approach to automated testing, observability, and structured logging?",
	}
}

func takeFirst(items []string, n int) []string {
	if len(items) <= n {
		return items
	}
	return items[:n]
}

// ---- Gemini Cloud AI Integration ----

func (s *AIService) geminiScoreCandidate(ctx context.Context, job *models.JobPosting, resumeText string, candidateSkills string) (*models.AIScoreResult, error) {
	prompt := fmt.Sprintf(`You are an expert technical recruiting AI. Analyze this candidate for the given job.
Return strictly valid JSON with this schema:
{
  "match_score": 85,
  "strengths": ["Skill1", "Skill2"],
  "missing_skills": ["Skill3"],
  "recommendation": "Strong Match",
  "summary_analysis": "Concise 2-sentence summary.",
  "interview_questions": ["Question 1", "Question 2", "Question 3"]
}

JOB TITLE: %s
REQUIRED SKILLS: %s
JOB REQUIREMENTS:
%s

CANDIDATE SKILLS: %s
RESUME CONTENT:
%s`, job.Title, job.RequiredSkills, job.Requirements, candidateSkills, resumeText)

	raw, err := s.callGemini(ctx, prompt)
	if err != nil {
		return nil, err
	}

	cleanJSON := cleanJSONBlock(raw)
	var res models.AIScoreResult
	if err := json.Unmarshal([]byte(cleanJSON), &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *AIService) geminiGenerateJD(ctx context.Context, req *models.GenerateJDRequest) (*models.GenerateJDResponse, error) {
	prompt := fmt.Sprintf(`Generate a professional, compelling job description for the following role:
Role: %s
Industry: %s
Experience Level: %s
Key Skills: %s
Workplace Type: %s

Return strictly valid JSON with keys:
{
  "title": "...",
  "description": "...",
  "requirements": "...",
  "required_skills": "comma separated skills"
}`, req.RoleTitle, req.Industry, req.ExperienceLevel, req.KeySkills, req.WorkplaceType)

	raw, err := s.callGemini(ctx, prompt)
	if err != nil {
		return nil, err
	}

	cleanJSON := cleanJSONBlock(raw)
	var res models.GenerateJDResponse
	if err := json.Unmarshal([]byte(cleanJSON), &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *AIService) geminiQuestions(ctx context.Context, job *models.JobPosting, resumeText string) ([]string, error) {
	prompt := fmt.Sprintf(`Generate 4 targeted, insightful technical & behavioral interview questions for a candidate applying for %s.
Candidate Resume: %s
Return strictly a JSON array of strings: ["Q1", "Q2", "Q3", "Q4"]`, job.Title, resumeText)

	raw, err := s.callGemini(ctx, prompt)
	if err != nil {
		return nil, err
	}
	cleanJSON := cleanJSONBlock(raw)
	var list []string
	if err := json.Unmarshal([]byte(cleanJSON), &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *AIService) callGemini(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=%s", s.apiKey)
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": prompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.2,
		},
	}
	data, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini api error: %s", string(bodyBytes))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", err
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", fmt.Errorf("empty gemini response")
}

func cleanJSONBlock(in string) string {
	s := strings.TrimSpace(in)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}
