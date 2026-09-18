package repository

import (
	"database/sql"
	"time"

	"talentpulse/internal/models"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

func (s *Store) InitSchema() error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id            BIGINT AUTO_INCREMENT PRIMARY KEY,
			email         VARCHAR(255) UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			full_name     VARCHAR(255) NOT NULL,
			role          VARCHAR(50) NOT NULL,
			company_name  VARCHAR(255),
			title         VARCHAR(255),
			bio           TEXT,
			skills        TEXT,
			resume_text   LONGTEXT,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS jobs (
			id               BIGINT AUTO_INCREMENT PRIMARY KEY,
			employer_id      BIGINT NOT NULL,
			employer_name    VARCHAR(255) NOT NULL,
			company_name     VARCHAR(255) NOT NULL,
			title            VARCHAR(255) NOT NULL,
			department       VARCHAR(255),
			location         VARCHAR(255) NOT NULL,
			workplace_type   VARCHAR(100) NOT NULL,
			employment_type  VARCHAR(100) NOT NULL,
			experience_level VARCHAR(100) NOT NULL,
			salary_min       BIGINT DEFAULT 0,
			salary_max       BIGINT DEFAULT 0,
			currency         VARCHAR(10) DEFAULT 'USD',
			description      LONGTEXT NOT NULL,
			requirements     LONGTEXT NOT NULL,
			required_skills  TEXT NOT NULL,
			status           VARCHAR(50) DEFAULT 'open',
			created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (employer_id) REFERENCES users(id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,

		`CREATE TABLE IF NOT EXISTS applications (
			id               BIGINT AUTO_INCREMENT PRIMARY KEY,
			job_id           BIGINT NOT NULL,
			candidate_id     BIGINT NOT NULL,
			resume_summary   TEXT,
			cover_letter     LONGTEXT,
			status           VARCHAR(50) DEFAULT 'applied',
			ai_match_score   INT DEFAULT 0,
			ai_strengths     TEXT,
			ai_gap_analysis  TEXT,
			ai_recommendation TEXT,
			notes            TEXT,
			created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			FOREIGN KEY (job_id) REFERENCES jobs(id),
			FOREIGN KEY (candidate_id) REFERENCES users(id),
			UNIQUE KEY uq_job_candidate (job_id, candidate_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`,
	}

	for _, stmt := range statements {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) CreateUser(u *models.User) error {
	query := `INSERT INTO users (email, password_hash, full_name, role, company_name, title, bio, skills, resume_text, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now().Format("2006-01-02 15:04:05")
	res, err := s.db.Exec(query, u.Email, u.PasswordHash, u.FullName, u.Role, u.CompanyName, u.Title, u.Bio, u.Skills, u.ResumeText, now)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	u.ID = id
	u.CreatedAt = now
	return nil
}

func (s *Store) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, password_hash, full_name, role, company_name, title, bio, skills, resume_text, created_at
		FROM users WHERE email = ?`
	row := s.db.QueryRow(query, email)
	var u models.User
	var comp, title, bio, skills, resume, createdAt sql.NullString
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &comp, &title, &bio, &skills, &resume, &createdAt)
	if err != nil {
		return nil, err
	}
	u.CompanyName = comp.String
	u.Title = title.String
	u.Bio = bio.String
	u.Skills = skills.String
	u.ResumeText = resume.String
	u.CreatedAt = createdAt.String
	return &u, nil
}

func (s *Store) GetUserByID(id int64) (*models.User, error) {
	query := `SELECT id, email, password_hash, full_name, role, company_name, title, bio, skills, resume_text, created_at
		FROM users WHERE id = ?`
	row := s.db.QueryRow(query, id)
	var u models.User
	var comp, title, bio, skills, resume, createdAt sql.NullString
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &comp, &title, &bio, &skills, &resume, &createdAt)
	if err != nil {
		return nil, err
	}
	u.CompanyName = comp.String
	u.Title = title.String
	u.Bio = bio.String
	u.Skills = skills.String
	u.ResumeText = resume.String
	u.CreatedAt = createdAt.String
	return &u, nil
}

func (s *Store) UpdateUserProfile(id int64, title, bio, skills, resumeText string) error {
	query := `UPDATE users SET title = ?, bio = ?, skills = ?, resume_text = ? WHERE id = ?`
	_, err := s.db.Exec(query, title, bio, skills, resumeText, id)
	return err
}

func (s *Store) CreateJob(j *models.JobPosting) error {
	query := `INSERT INTO jobs (employer_id, employer_name, company_name, title, department, location,
		workplace_type, employment_type, experience_level, salary_min, salary_max, currency,
		description, requirements, required_skills, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now().Format("2006-01-02 15:04:05")
	res, err := s.db.Exec(query, j.EmployerID, j.EmployerName, j.CompanyName, j.Title, j.Department, j.Location,
		j.WorkplaceType, j.EmploymentType, j.ExperienceLevel, j.SalaryMin, j.SalaryMax, j.Currency,
		j.Description, j.Requirements, j.RequiredSkills, "open", now, now)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	j.ID = id
	j.Status = "open"
	j.CreatedAt = now
	j.UpdatedAt = now
	return nil
}

func (s *Store) ListJobs(search, workplaceType, expLevel string) ([]models.JobPosting, error) {
	query := `SELECT j.id, j.employer_id, j.employer_name, j.company_name, j.title, j.department, j.location,
		j.workplace_type, j.employment_type, j.experience_level, j.salary_min, j.salary_max, j.currency,
		j.description, j.requirements, j.required_skills, j.status, j.created_at, j.updated_at,
		(SELECT COUNT(*) FROM applications a WHERE a.job_id = j.id) as applicant_count
		FROM jobs j WHERE j.status = 'open'`

	var args []interface{}
	if search != "" {
		query += " AND (j.title LIKE ? OR j.company_name LIKE ? OR j.required_skills LIKE ?)"
		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern)
	}
	if workplaceType != "" && workplaceType != "all" {
		query += " AND LOWER(j.workplace_type) = LOWER(?)"
		args = append(args, workplaceType)
	}
	if expLevel != "" && expLevel != "all" {
		query += " AND LOWER(j.experience_level) = LOWER(?)"
		args = append(args, expLevel)
	}
	query += " ORDER BY j.created_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobs := []models.JobPosting{}
	for rows.Next() {
		var j models.JobPosting
		var dept, createdAt, updatedAt sql.NullString
		err := rows.Scan(&j.ID, &j.EmployerID, &j.EmployerName, &j.CompanyName, &j.Title, &dept, &j.Location,
			&j.WorkplaceType, &j.EmploymentType, &j.ExperienceLevel, &j.SalaryMin, &j.SalaryMax, &j.Currency,
			&j.Description, &j.Requirements, &j.RequiredSkills, &j.Status, &createdAt, &updatedAt, &j.ApplicantCount)
		if err != nil {
			return nil, err
		}
		j.Department = dept.String
		j.CreatedAt = createdAt.String
		j.UpdatedAt = updatedAt.String
		jobs = append(jobs, j)
	}
	return jobs, nil
}

func (s *Store) GetJobByID(id int64) (*models.JobPosting, error) {
	query := `SELECT j.id, j.employer_id, j.employer_name, j.company_name, j.title, j.department, j.location,
		j.workplace_type, j.employment_type, j.experience_level, j.salary_min, j.salary_max, j.currency,
		j.description, j.requirements, j.required_skills, j.status, j.created_at, j.updated_at,
		(SELECT COUNT(*) FROM applications a WHERE a.job_id = j.id) as applicant_count
		FROM jobs j WHERE j.id = ?`

	row := s.db.QueryRow(query, id)
	var j models.JobPosting
	var dept, createdAt, updatedAt sql.NullString
	err := row.Scan(&j.ID, &j.EmployerID, &j.EmployerName, &j.CompanyName, &j.Title, &dept, &j.Location,
		&j.WorkplaceType, &j.EmploymentType, &j.ExperienceLevel, &j.SalaryMin, &j.SalaryMax, &j.Currency,
		&j.Description, &j.Requirements, &j.RequiredSkills, &j.Status, &createdAt, &updatedAt, &j.ApplicantCount)
	if err != nil {
		return nil, err
	}
	j.Department = dept.String
	j.CreatedAt = createdAt.String
	j.UpdatedAt = updatedAt.String
	return &j, nil
}

func (s *Store) CreateApplication(app *models.Application) error {
	query := `INSERT INTO applications (job_id, candidate_id, resume_summary, cover_letter, status,
		ai_match_score, ai_strengths, ai_gap_analysis, ai_recommendation, notes, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now().Format("2006-01-02 15:04:05")
	res, err := s.db.Exec(query, app.JobID, app.CandidateID, app.ResumeSummary, app.CoverLetter, models.AppStatusApplied,
		app.AIMatchScore, app.AIStrengths, app.AIGapAnalysis, app.AIRecommendation, app.Notes, now, now)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	app.ID = id
	app.Status = models.AppStatusApplied
	app.CreatedAt = now
	app.UpdatedAt = now
	return nil
}

func (s *Store) GetApplicationsByJob(jobID int64) ([]models.Application, error) {
	query := `SELECT a.id, a.job_id, a.candidate_id, u.full_name, u.email, u.skills,
		a.resume_summary, a.cover_letter, a.status, a.ai_match_score, a.ai_strengths, a.ai_gap_analysis, a.ai_recommendation,
		COALESCE(a.notes, ''), a.created_at, a.updated_at
		FROM applications a
		JOIN users u ON a.candidate_id = u.id
		WHERE a.job_id = ?
		ORDER BY a.ai_match_score DESC, a.created_at DESC`

	rows, err := s.db.Query(query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []models.Application
	for rows.Next() {
		var a models.Application
		var notes, createdAt, updatedAt sql.NullString
		err := rows.Scan(&a.ID, &a.JobID, &a.CandidateID, &a.CandidateName, &a.CandidateEmail, &a.CandidateSkills,
			&a.ResumeSummary, &a.CoverLetter, &a.Status, &a.AIMatchScore, &a.AIStrengths, &a.AIGapAnalysis, &a.AIRecommendation,
			&notes, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		a.Notes = notes.String
		a.CreatedAt = createdAt.String
		a.UpdatedAt = updatedAt.String
		apps = append(apps, a)
	}
	return apps, nil
}

func (s *Store) GetApplicationsByCandidate(candidateID int64) ([]models.Application, error) {
	query := `SELECT a.id, a.job_id, j.title, j.company_name, a.candidate_id,
		a.resume_summary, a.cover_letter, a.status, a.ai_match_score, a.ai_strengths, a.ai_gap_analysis, a.ai_recommendation,
		COALESCE(a.notes, ''), a.created_at, a.updated_at
		FROM applications a
		JOIN jobs j ON a.job_id = j.id
		WHERE a.candidate_id = ?
		ORDER BY a.created_at DESC`

	rows, err := s.db.Query(query, candidateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var apps []models.Application
	for rows.Next() {
		var a models.Application
		var notes, createdAt, updatedAt sql.NullString
		err := rows.Scan(&a.ID, &a.JobID, &a.JobTitle, &a.CompanyName, &a.CandidateID,
			&a.ResumeSummary, &a.CoverLetter, &a.Status, &a.AIMatchScore, &a.AIStrengths, &a.AIGapAnalysis, &a.AIRecommendation,
			&notes, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		a.Notes = notes.String
		a.CreatedAt = createdAt.String
		a.UpdatedAt = updatedAt.String
		apps = append(apps, a)
	}
	return apps, nil
}

func (s *Store) UpdateApplicationStatus(appID int64, status models.ApplicationStatus, notes string) error {
	query := `UPDATE applications SET status = ?, notes = ?, updated_at = ? WHERE id = ?`
	_, err := s.db.Exec(query, status, notes, time.Now().Format("2006-01-02 15:04:05"), appID)
	return err
}
