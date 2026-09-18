package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/go-sql-driver/mysql"

	"talentpulse/internal/ai"
	"talentpulse/internal/handlers"
	"talentpulse/internal/middleware"
	"talentpulse/internal/models"
	"talentpulse/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	port := getEnv("PORT", "8080")

	// MySQL connection parameters (read from env with defaults)
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASSWORD", "osl20031975@0210")
	dbName := getEnv("DB_NAME", "talent_pulse_db")

	// First connect WITHOUT specifying a database to create it if needed
	rootDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		dbUser, dbPass, dbHost, dbPort)
	rootDB, err := sql.Open("mysql", rootDSN)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MySQL: %v", err)
	}
	_, err = rootDB.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName))
	if err != nil {
		log.Fatalf("❌ Failed to create database '%s': %v", dbName, err)
	}
	rootDB.Close()

	// Now connect with the target database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
		dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ Failed to open database '%s': %v", dbName, err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("❌ Cannot reach MySQL at %s:%s — %v", dbHost, dbPort, err)
	}
	fmt.Printf("✅ Connected to MySQL database '%s' at %s:%s\n", dbName, dbHost, dbPort)

	store := repository.NewStore(db)
	if err := store.InitSchema(); err != nil {
		log.Fatalf("❌ Failed to initialize database schema: %v", err)
	}

	seedInitialData(store)

	aiService := ai.NewAIService()
	h := handlers.NewHandler(store, aiService)

	r := chi.NewRouter()
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.CORS)

	// API Routes
	r.Route("/api", func(api chi.Router) {
		// Public Auth
		api.Post("/auth/register", h.Register)
		api.Post("/auth/login", h.Login)

		// Public Job discovery
		api.Get("/jobs", h.ListJobs)
		api.Get("/jobs/{id}", h.GetJob)

		// AI Assistant (Public or preview)
		api.Post("/ai/generate-jd", h.AIGenerateJD)
		api.Post("/jobs/{id}/ai-match-preview", h.AIMatchPreview)
		api.Post("/jobs/{id}/ai-questions", h.AIInterviewQuestions)

		// Authenticated Routes
		api.Group(func(auth chi.Router) {
			auth.Use(middleware.AuthMiddleware)

			auth.Get("/auth/me", h.GetMe)
			auth.Put("/auth/me", h.UpdateMe)

			// Candidate endpoints
			auth.Post("/jobs/{id}/apply", h.ApplyJob)
			auth.Get("/candidate/applications", h.GetMyApplications)

			// Employer endpoints
			auth.Post("/jobs", h.CreateJob)
			auth.Get("/jobs/{id}/applicants", h.GetJobApplicants)
			auth.Put("/applications/{appId}/status", h.UpdateApplicationStatus)
		})
	})

	// Static UI file server
	staticDir := http.Dir("./web/static")
	fileServer := http.FileServer(staticDir)
	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			http.NotFound(w, r)
			return
		}
		f, err := staticDir.Open(r.URL.Path)
		if err == nil {
			f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, "./web/static/index.html")
	}))

	fmt.Printf("\n🚀 TalentPulse AI Job Portal Server started on http://localhost:%s\n\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func seedInitialData(store *repository.Store) {
	// Check if already seeded
	_, err := store.GetUserByEmail("recruiter@google.com")
	if err == nil {
		return // already seeded
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	hashStr := string(hash)

	// ─── Recruiters ─────────────────────────────────────────────────────────────

	google := &models.User{
		Email:        "recruiter@google.com",
		PasswordHash: hashStr,
		FullName:     "Sarah De Silva",
		Role:         models.RoleEmployer,
		CompanyName:  "Google DeepMind",
		Title:        "Principal Technical Talent Partner",
		Bio:          "Recruiting world-class engineers for Google DeepMind and Google Cloud AI infrastructure teams. Passionate about matching talent to high-impact problems.",
	}
	_ = store.CreateUser(google)

	meta := &models.User{
		Email:        "recruiter@meta.com",
		PasswordHash: hashStr,
		FullName:     "James Holloway",
		Role:         models.RoleEmployer,
		CompanyName:  "Meta Platforms",
		Title:        "Senior Engineering Recruiter",
		Bio:          "Talent acquisition lead for Meta's Reality Labs and Infrastructure Engineering. Building teams that ship at billions-user scale.",
	}
	_ = store.CreateUser(meta)

	amazon := &models.User{
		Email:        "recruiter@amazon.com",
		PasswordHash: hashStr,
		FullName:     "Priya Nair",
		Role:         models.RoleEmployer,
		CompanyName:  "Amazon Web Services",
		Title:        "AWS Sr. Tech Recruiting Manager",
		Bio:          "Scaling AWS engineering orgs across cloud infrastructure, serverless, and developer tooling. 10+ years in deep tech talent.",
	}
	_ = store.CreateUser(amazon)

	microsoft := &models.User{
		Email:        "recruiter@microsoft.com",
		PasswordHash: hashStr,
		FullName:     "Daniel Osei",
		Role:         models.RoleEmployer,
		CompanyName:  "Microsoft Azure",
		Title:        "Principal Recruiter — Azure Engineering",
		Bio:          "Hiring engineers and architects for Microsoft Azure's core platform, AI services, and developer experience teams globally.",
	}
	_ = store.CreateUser(microsoft)

	// ─── Candidates ─────────────────────────────────────────────────────────────

	alex := &models.User{
		Email:        "anne.dev@gmail.com",
		PasswordHash: hashStr,
		FullName:     "Anne Devara",
		Role:         models.RoleCandidate,
		Title:        "Senior Go & Cloud Systems Engineer",
		Bio:          "Passionate about building scalable backend microservices, event streaming, and resilient APIs.",
		Skills:       "Go, Golang, Kubernetes, Docker, PostgreSQL, Redis, gRPC, Microservices, CI/CD, AWS",
		ResumeText:   "Alex Mercer — Senior Software Engineer with 6+ years of production experience in Go (Golang), distributed caching with Redis, high-throughput microservices using gRPC/Protobuf, PostgreSQL schema design, and Docker/Kubernetes deployments on AWS. Led platform migrations serving 50M+ daily users.",
	}
	_ = store.CreateUser(alex)

	nina := &models.User{
		Email:        "nina.chen@outlook.com",
		PasswordHash: hashStr,
		FullName:     "Nina Chen",
		Role:         models.RoleCandidate,
		Title:        "ML Engineer & AI Researcher",
		Bio:          "Combining cutting-edge ML research with production-grade Python systems. Published at NeurIPS and ICML.",
		Skills:       "Python, PyTorch, TensorFlow, LLMs, Transformers, MLflow, Kubernetes, CUDA, Data Pipelines, AWS SageMaker",
		ResumeText:   "Nina Chen — ML Engineer with 5 years building large-scale recommendation systems, fine-tuned transformer models, and deploying ML pipelines on Kubernetes. MS in Computer Science (Stanford). Patent holder in federated learning.",
	}
	_ = store.CreateUser(nina)

	// ─── Jobs ────────────────────────────────────────────────────────────────────

	jobs := []models.JobPosting{
		// Google DeepMind
		{
			EmployerID:      google.ID,
			EmployerName:    google.FullName,
			CompanyName:     "Google DeepMind",
			Title:           "Staff Research Engineer — Distributed AI Systems",
			Department:      "Core AI Infrastructure",
			Location:        "London, UK / Remote",
			WorkplaceType:   "Remote",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Senior",
			SalaryMin:       200000,
			SalaryMax:       280000,
			Currency:        "USD",
			Description:     "Design and build the foundational distributed training infrastructure powering Google DeepMind's next-generation models including Gemini. You will architect petabyte-scale data pipelines, optimize GPU/TPU cluster scheduling, and collaborate directly with leading AI researchers to translate bleeding-edge research into reliable production systems.",
			Requirements:    "- 7+ years of production engineering experience, with 3+ years on distributed ML infrastructure.\n- Expert-level proficiency in Python and/or Go for systems programming.\n- Deep understanding of ML frameworks (JAX, PyTorch) and hardware accelerators (TPU/GPU).\n- Experience with large-scale job schedulers (Borg, SLURM, Ray).\n- Publications or patents in distributed systems or ML infrastructure are a plus.",
			RequiredSkills:  "Python, Go, JAX, PyTorch, Kubernetes, gRPC, Distributed Systems, TPU, CUDA, Protobuf",
		},
		{
			EmployerID:      google.ID,
			EmployerName:    google.FullName,
			CompanyName:     "Google Cloud",
			Title:           "Senior Site Reliability Engineer — Cloud Spanner",
			Department:      "Cloud Databases SRE",
			Location:        "Seattle, WA / Hybrid",
			WorkplaceType:   "Hybrid",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Senior",
			SalaryMin:       175000,
			SalaryMax:       245000,
			Currency:        "USD",
			Description:     "Own the reliability, scalability, and performance of Cloud Spanner — the world's first globally distributed relational database. Drive SLO engineering, automate toil, conduct chaos engineering experiments, and lead incident response for a service trusted by thousands of enterprise customers.",
			Requirements:    "- 5+ years SRE or platform engineering experience.\n- Strong Go or Python for automation and tooling.\n- Hands-on experience with distributed databases (Spanner, CockroachDB, Cassandra).\n- Expertise in observability stacks (Prometheus, OpenTelemetry, Grafana).\n- Proven incident management and blameless post-mortem culture.",
			RequiredSkills:  "Go, Python, Kubernetes, Spanner, PostgreSQL, Prometheus, Terraform, Linux, SRE",
		},
		{
			EmployerID:      google.ID,
			EmployerName:    google.FullName,
			CompanyName:     "Google Security",
			Title:           "Software Engineer — Cloud Security & Zero Trust",
			Department:      "Identity & Access Security",
			Location:        "Sunnyvale, CA / On-site",
			WorkplaceType:   "On-site",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Mid-Senior",
			SalaryMin:       165000,
			SalaryMax:       230000,
			Currency:        "USD",
			Description:     "Build the next generation of Google's Zero Trust Network Access (BeyondCorp) platform. Develop identity-aware proxies, cryptographic attestation pipelines, and policy-enforcement engines that protect Google's internal infrastructure and are offered to GCP enterprise customers.",
			Requirements:    "- 4+ years of security engineering or backend systems experience.\n- Proficiency in Go, C++, or Java for security-critical services.\n- Knowledge of PKI, mTLS, OAuth 2.0, SPIFFE/SPIRE, and zero-trust architectures.\n- Experience with cryptography libraries and secure code review practices.",
			RequiredSkills:  "Go, Security, Zero Trust, OAuth2, mTLS, Cryptography, Kubernetes, Linux, BeyondCorp",
		},

		// Meta Platforms
		{
			EmployerID:      meta.ID,
			EmployerName:    meta.FullName,
			CompanyName:     "Meta AI",
			Title:           "Research Engineer — Large Language Models",
			Department:      "FAIR (Fundamental AI Research)",
			Location:        "Menlo Park, CA / Hybrid",
			WorkplaceType:   "Hybrid",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Senior",
			SalaryMin:       210000,
			SalaryMax:       300000,
			Currency:        "USD",
			Description:     "Join Meta FAIR to push the frontier of open-source large language models (Llama series). You will co-design model architectures, implement efficient training kernels, run large-scale ablations across thousands of GPUs, and contribute to publications that shape the global AI research community.",
			Requirements:    "- PhD or equivalent research experience in ML, NLP, or AI.\n- Strong Python and CUDA/Triton kernel programming.\n- Experience training models with 70B+ parameters.\n- Track record of top-tier publications (NeurIPS, ICML, ACL, ICLR).\n- Familiarity with RLHF, DPO, and instruction-tuning pipelines.",
			RequiredSkills:  "Python, PyTorch, CUDA, Triton, LLMs, Transformers, RLHF, Distributed Training, NCCL",
		},
		{
			EmployerID:      meta.ID,
			EmployerName:    meta.FullName,
			CompanyName:     "Meta Platforms",
			Title:           "Production Engineer — Infrastructure Reliability",
			Department:      "Infrastructure Engineering",
			Location:        "Remote — US",
			WorkplaceType:   "Remote",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Mid-Senior",
			SalaryMin:       155000,
			SalaryMax:       220000,
			Currency:        "USD",
			Description:     "Meta's Production Engineers sit at the intersection of software engineering and systems operations. Own the reliability of Meta's core social graph infrastructure serving 3 billion daily active users. Drive capacity planning, performance profiling, and automated failure recovery systems for services running on one of the world's largest private compute fleets.",
			Requirements:    "- 4+ years of systems engineering or SRE experience.\n- Proficiency in C++, Python, or Hack/PHP for infrastructure tooling.\n- Deep knowledge of Linux kernel internals, networking, and storage systems.\n- Experience with high-throughput caching layers (Memcache, TAO, CacheLib).",
			RequiredSkills:  "C++, Python, Linux, Networking, Caching, Kubernetes, Observability, Capacity Planning",
		},
		{
			EmployerID:      meta.ID,
			EmployerName:    meta.FullName,
			CompanyName:     "Meta Reality Labs",
			Title:           "Senior Software Engineer — AR/VR Rendering",
			Department:      "Reality Labs — Graphics Platform",
			Location:        "Redmond, WA / Hybrid",
			WorkplaceType:   "Hybrid",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Senior",
			SalaryMin:       190000,
			SalaryMax:       265000,
			Currency:        "USD",
			Description:     "Define the visual fidelity of the next generation of augmented and virtual reality experiences for Meta Quest and Ray-Ban smart glasses. Build high-performance real-time rendering engines, optimize shader pipelines, and collaborate with hardware teams to extract maximum performance from custom silicon.",
			Requirements:    "- 5+ years of graphics programming with C++ and OpenGL/Vulkan/Metal.\n- Expert knowledge of PBR shading models, shadow algorithms, and post-processing.\n- Experience optimizing for mobile/embedded GPU architectures.\n- Familiarity with AR/VR-specific techniques: reprojection, foveated rendering, hand tracking.",
			RequiredSkills:  "C++, Vulkan, OpenGL, GLSL, Real-time Rendering, GPU Optimization, AR/VR, Computer Vision",
		},

		// Amazon Web Services
		{
			EmployerID:      amazon.ID,
			EmployerName:    amazon.FullName,
			CompanyName:     "Amazon Web Services",
			Title:           "Principal Engineer — Serverless Compute (Lambda)",
			Department:      "AWS Serverless",
			Location:        "Seattle, WA / Hybrid",
			WorkplaceType:   "Hybrid",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Lead",
			SalaryMin:       220000,
			SalaryMax:       320000,
			Currency:        "USD",
			Description:     "Drive the technical vision and multi-year roadmap for AWS Lambda — the world's most widely used serverless compute platform. As a Principal Engineer, you will influence the architecture of the entire AWS serverless portfolio, mentor senior engineers, set engineering standards, and represent AWS at industry conferences and standards bodies.",
			Requirements:    "- 12+ years of software engineering with 5+ years at senior/principal level.\n- Expert in distributed systems, OS internals (Linux), and container/VM isolation.\n- Deep experience with high-scale microservices (millions of RPS).\n- Track record of shaping org-wide technical strategy and culture.\n- Strong written communication — able to author compelling 6-pager documents.",
			RequiredSkills:  "Go, Rust, C++, Linux, Containers, Serverless, Distributed Systems, AWS, Architecture",
		},
		{
			EmployerID:      amazon.ID,
			EmployerName:    amazon.FullName,
			CompanyName:     "Amazon Web Services",
			Title:           "Senior Data Engineer — Real-Time Analytics Platform",
			Department:      "AWS Analytics",
			Location:        "Remote — US / Europe",
			WorkplaceType:   "Remote",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Senior",
			SalaryMin:       150000,
			SalaryMax:       210000,
			Currency:        "USD",
			Description:     "Build the data infrastructure powering Amazon's real-time analytics products (Kinesis, Redshift, OpenSearch). Design and implement petabyte-scale ETL pipelines, streaming data architectures, and self-service data platforms used by thousands of internal and external AWS teams.",
			Requirements:    "- 5+ years of data engineering experience.\n- Expert in Apache Kafka, Apache Flink, or Spark Streaming.\n- Strong SQL and Python or Java for pipeline development.\n- Experience with data lake architectures (Delta Lake, Apache Iceberg, Hudi).\n- Knowledge of AWS analytics services (Redshift, Glue, EMR, Kinesis).",
			RequiredSkills:  "Python, SQL, Kafka, Flink, Spark, Redshift, S3, AWS Glue, Data Lake, Airflow",
		},
		{
			EmployerID:      amazon.ID,
			EmployerName:    amazon.FullName,
			CompanyName:     "Amazon",
			Title:           "Applied Scientist — Personalization & Recommendations",
			Department:      "Amazon Personalize",
			Location:        "New York, NY / Remote",
			WorkplaceType:   "Remote",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Senior",
			SalaryMin:       185000,
			SalaryMax:       260000,
			Currency:        "USD",
			Description:     "Own the science and implementation of next-generation recommendation algorithms that personalize shopping experiences for hundreds of millions of Amazon customers globally. Research novel approaches to collaborative filtering, session-based recommendations, multi-objective ranking, and causal inference at Amazon-scale.",
			Requirements:    "- PhD in ML, Statistics, Operations Research, or related field.\n- 3+ years applied experience building production recommendation/ranking systems.\n- Proficiency in Python with PyTorch or TensorFlow.\n- Experience with A/B testing frameworks and causal inference methods.\n- Strong publication record or applied patent portfolio.",
			RequiredSkills:  "Python, PyTorch, Recommender Systems, A/B Testing, Causal Inference, Spark, AWS SageMaker",
		},

		// Microsoft Azure
		{
			EmployerID:      microsoft.ID,
			EmployerName:    microsoft.FullName,
			CompanyName:     "Microsoft Azure",
			Title:           "Principal Software Engineer — Azure Kubernetes Service",
			Department:      "Azure Container Platform",
			Location:        "Redmond, WA / Remote",
			WorkplaceType:   "Remote",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Lead",
			SalaryMin:       195000,
			SalaryMax:       280000,
			Currency:        "USD",
			Description:     "Lead the engineering of Azure Kubernetes Service (AKS), Microsoft's managed Kubernetes offering used by enterprises worldwide. Drive key investments in node auto-provisioning, multi-tenancy security hardening, Windows container support, and the AKS control plane that manages millions of cluster-hours daily.",
			Requirements:    "- 10+ years software engineering; 4+ years on Kubernetes or container platforms.\n- Deep expertise in Go (primary language of Kubernetes ecosystem).\n- Experience contributing to upstream Kubernetes or CNCF projects.\n- Strong knowledge of Linux networking (CNI, eBPF, service meshes).\n- Ability to drive technical decisions across multiple teams.",
			RequiredSkills:  "Go, Kubernetes, Docker, eBPF, CNI, Azure, Helm, Terraform, Linux, Networking",
		},
		{
			EmployerID:      microsoft.ID,
			EmployerName:    microsoft.FullName,
			CompanyName:     "Microsoft",
			Title:           "Senior Software Engineer — GitHub Copilot",
			Department:      "GitHub Next",
			Location:        "San Francisco, CA / Remote",
			WorkplaceType:   "Remote",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Senior",
			SalaryMin:       170000,
			SalaryMax:       245000,
			Currency:        "USD",
			Description:     "Shape the future of AI-powered developer tooling as part of the GitHub Copilot engineering team. Build the backend systems powering code completion, code explanation, PR summaries, and Copilot Chat. Work at the intersection of language model serving, developer experience, and IDE integration across VS Code, JetBrains, Neovim, and GitHub.com.",
			Requirements:    "- 5+ years backend engineering with TypeScript/Node.js or Go.\n- Experience integrating and serving large language model APIs at scale.\n- Knowledge of IDE extension development (VS Code API, LSP protocol).\n- Understanding of developer workflows: Git, CI/CD, code review, testing.\n- Experience with low-latency streaming APIs (SSE, WebSockets).",
			RequiredSkills:  "TypeScript, Go, Node.js, LLMs, REST APIs, VS Code, LSP, Git, GitHub Actions, Azure",
		},
		{
			EmployerID:      microsoft.ID,
			EmployerName:    microsoft.FullName,
			CompanyName:     "Microsoft Azure",
			Title:           "Cloud Solutions Architect — Enterprise AI",
			Department:      "Azure Customer Success Engineering",
			Location:        "Remote — Global",
			WorkplaceType:   "Remote",
			EmploymentType:  "Full-time",
			ExperienceLevel: "Senior",
			SalaryMin:       145000,
			SalaryMax:       200000,
			Currency:        "USD",
			Description:     "Partner with Fortune 500 enterprise customers to architect and deploy AI-powered solutions on Azure OpenAI Service, Azure AI Studio, and Cognitive Services. Translate complex business requirements into scalable cloud architectures, run proof-of-concept engagements, and deliver best-practice design patterns that accelerate AI adoption.",
			Requirements:    "- 6+ years of cloud architecture or solutions engineering experience.\n- Azure certifications (Solutions Architect Expert or AI Engineer Associate preferred).\n- Hands-on experience with Azure OpenAI, Azure AI Search, and RAG patterns.\n- Strong communication skills — able to present to C-level executives.\n- Python or TypeScript for building demonstration solutions.",
			RequiredSkills:  "Azure, Python, TypeScript, Azure OpenAI, RAG, Cognitive Services, Terraform, Enterprise Architecture",
		},
	}

	for i := range jobs {
		_ = store.CreateJob(&jobs[i])
	}
}

