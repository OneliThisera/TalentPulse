# TalentPulse AI - Real-World Job Portal System (Go + AI)

TalentPulse AI is an enterprise-grade job portal backend and portal system built with **Go (Golang)**, featuring native AI candidate evaluation, resume fit scoring, AI Job Description generation, and candidate screening interview question formulation.

## 🚀 Key Features

1. **AI Resume Match & Fit Scoring**:
   - Compares candidate resume and declared skills against required competencies.
   - Calculates a 0-100% Fit Score, categorizes candidate recommendations (Strong Match, Moderate, Low), and pinpoints missing skills.
   - Generates customized technical interview screening questions tailored to the candidate's gaps and strengths.

2. **AI Job Description Generator**:
   - Recruiters can generate structured, high-converting job postings with rich responsibilities, technical requirements, and skills in seconds.

3. **Recruiter Hiring Pipeline**:
   - Employers can inspect candidate applications ranked by AI match score.
   - Move applicants through pipeline stages (`Applied` -> `Reviewing` -> `Shortlisted` -> `Interview Scheduled` -> `Offer Extended` -> `Rejected`).

4. **Candidate Career Discovery**:
   - Interactive job search with real-time filters (Workplace type: Remote/Hybrid/On-site, Experience level: Junior/Mid-Senior/Senior/Lead, Keyword search).
   - "⚡ AI Match" preview directly from job cards before applying.
   - Application status tracking portal.

5. **Clean Go Architecture**:
   - `cmd/server/main.go`: Server bootstrap and seed data.
   - `internal/models/`: Domain models and payload schemas.
   - `internal/handlers/`: High-performance RESTful API endpoints.
   - `internal/repository/`: Pure Go SQLite persistence layer (zero external C dependencies).
   - `internal/ai/`: Modular AI orchestration (supports Google Gemini / OpenAI Cloud LLM API with intelligent fallback local engine).
   - `internal/middleware/`: JWT authentication, RBAC, CORS, and logging.

---

## 🏃 Running the Application

### 1. Start Server
```powershell
cd f:\Projects\talentpulse
.\talentpulse.exe
```

The portal server will be live at:
👉 **http://localhost:8080**

### 2. Demo User Credentials (Pre-seeded)

| Role | Email | Password |
|---|---|---|
| **Recruiter / Employer** | `recruiter@google.com` | `password123` |
| **Candidate / Job Seeker** | `alex.dev@gmail.com` | `password123` |

You can also register brand new candidate or employer accounts directly through the UI modal.

---

## 🤖 AI Configuration (Optional)
To use Google Gemini Cloud LLM for live generative AI:
Set your environment variable:
```powershell
$env:GEMINI_API_KEY="your-gemini-api-key"
```
If no key is supplied, TalentPulse automatically runs its built-in local heuristic NLP engine so all AI matching and JD generation features work immediately offline.
