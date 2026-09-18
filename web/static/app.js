// TalentPulse SPA Client
const state = {
  user: JSON.parse(localStorage.getItem("tp_user") || "null"),
  token: localStorage.getItem("tp_token") || "",
  jobs: [],
  activeView: "browse",
  filter: { search: "", workplace: "all", experience: "all" }
};
const api = {
  async req(url, options = {}) {
    const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
    if (state.token) headers["Authorization"] = "Bearer " + state.token;
    const res = await fetch(url, { ...options, headers });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.error || "Server error");
    return data;
  }
};
const app = {
  init() { this.renderNavAuth(); this.navigate("browse"); },
  toast(msg, type = "info") {
    const box = document.getElementById("toast-container");
    const el = document.createElement("div");
    el.className = "toast";
    const icon = type === "error" ? "❌" : (type === "success" ? "✅" : "⚡");
    el.innerHTML = "<span>" + icon + "</span><span>" + msg + "</span>";
    box.appendChild(el);
    setTimeout(() => el.remove(), 4000);
  },
  renderNavAuth() {
    const container = document.getElementById("nav-auth-section");
    if (!state.user) {
      container.innerHTML = `<button class="btn btn-outline btn-sm" onclick="app.showAuthModal('login')">Sign In</button><button class="btn btn-primary btn-sm" onclick="app.showAuthModal('register')">Get Started</button>`;
    } else {
      const color = state.user.role === "employer" ? "#ec4899" : "#6366f1";
      container.innerHTML = `<div style="display:flex; align-items:center; gap:0.75rem;"><div style="text-align:right;"><div style="font-weight:700; font-size:0.9rem;">${state.user.full_name}</div><div style="font-size:0.75rem; color:${color}; text-transform:uppercase; font-weight:700;">${state.user.role}</div></div><button class="btn btn-outline btn-sm" onclick="app.logout()">Logout</button></div>`;
    }
  },
  logout() {
    state.user = null; state.token = ""; localStorage.removeItem("tp_user"); localStorage.removeItem("tp_token");
    this.renderNavAuth(); this.toast("Signed out", "info"); this.navigate("browse");
  },
  navigate(view) {
    state.activeView = view;
    document.querySelectorAll(".nav-btn").forEach(btn => btn.classList.remove("active"));
    const b = document.getElementById("nav-" + view); if (b) b.classList.add("active");
    const m = document.getElementById("main-content");
    if (view === "browse") this.renderBrowse(m);
    else if (view === "post-job") this.renderPostJob(m);
    else if (view === "dashboard") this.renderRecruiterDashboard(m);
    else if (view === "my-apps") this.renderCandidateApps(m);
  },
  async renderBrowse(c) {
    c.innerHTML = `
      <section class="hero-banner">
        <div class="hero-pill">✨ Powered by Go &amp; AI Candidate Fit Scoring</div>
        <h1 class="hero-title">Hire Faster with <span>AI Matching</span> &amp; Precision</h1>
        <p class="hero-desc">Discover high-impact engineering roles with automated skill gap analysis, AI screening, and instant match scoring.</p>
        <div class="search-box-wrap">
          <input type="text" id="search-keyword" class="search-input" placeholder="Search by title, company, or skill (e.g. Go, Kubernetes)..." />
          <select id="filter-workplace" class="filter-select">
            <option value="all">All Workplaces</option><option value="Remote">Remote</option><option value="Hybrid">Hybrid</option><option value="On-site">On-site</option>
          </select>
          <select id="filter-experience" class="filter-select">
            <option value="all">All Experience</option><option value="Junior">Junior</option><option value="Mid-Senior">Mid-Senior</option><option value="Senior">Senior</option><option value="Lead">Lead</option>
          </select>
          <button class="btn btn-primary" onclick="app.searchJobs()">Search</button>
        </div>
      </section>
      <div id="job-list-container" class="job-grid"><div style="color:var(--text-muted); text-align:center; grid-column:1/-1;">Loading jobs...</div></div>
    `;
    document.getElementById("search-keyword").addEventListener("keyup", (e) => { if (e.key === "Enter") this.searchJobs(); });
    await this.fetchJobs();
  },
  async fetchJobs() {
    try {
      const q = new URLSearchParams({ q: state.filter.search, workplace: state.filter.workplace, experience: state.filter.experience });
      state.jobs = await api.req("/api/jobs?" + q.toString());
      this.renderJobList(state.jobs);
    } catch (err) { this.toast(err.message, "error"); }
  },
  searchJobs() {
    state.filter.search = document.getElementById("search-keyword").value;
    state.filter.workplace = document.getElementById("filter-workplace").value;
    state.filter.experience = document.getElementById("filter-experience").value;
    this.fetchJobs();
  },
  renderJobList(jobs) {
    const el = document.getElementById("job-list-container");
    if (!jobs || !jobs.length) { el.innerHTML = `<div style="grid-column:1/-1; text-align:center; padding:3rem; color:var(--text-muted);">No open roles found.</div>`; return; }
    el.innerHTML = jobs.map(j => `
      <div class="job-card">
        <div class="job-header">
          <div><div class="company-badge">${j.company_name}</div><h3 class="job-title">${j.title}</h3></div>
          <span class="tag tag-accent">${j.workplace_type}</span>
        </div>
        <p style="color:var(--text-muted); font-size:0.92rem; line-height:1.5; margin:0.5rem 0 1rem; display:-webkit-box; -webkit-line-clamp:3; -webkit-box-orient:vertical; overflow:hidden;">${j.description}</p>
        <div class="tag-list">
          ${j.required_skills.split(",").slice(0, 4).map(s => `<span class="tag">${s.trim()}</span>`).join("")}
          <span class="tag">${j.experience_level}</span><span class="tag">📍 ${j.location}</span>
        </div>
        <div class="job-salary">
          <div>$${(j.salary_min/1000).toFixed(0)}k - $${(j.salary_max/1000).toFixed(0)}k <span style="font-size:0.75rem; color:var(--text-dim);">/yr</span></div>
          <div style="display:flex; gap:0.5rem;">
            <button class="btn btn-outline btn-sm" onclick="app.previewAIMatch(${j.id})">⚡ AI Match</button>
            <button class="btn btn-primary btn-sm" onclick="app.openApplyModal(${j.id})">Apply</button>
          </div>
        </div>
      </div>
    `).join("");
  },
  closeModal() { document.getElementById("modal-container").classList.add("hidden"); },
  async previewAIMatch(jobID) {
    const job = state.jobs.find(j => j.id === jobID); if (!job) return;
    const modal = document.getElementById("modal-content");
    modal.innerHTML = `
      <button class="modal-close" onclick="app.closeModal()">&times;</button>
      <h2 style="font-size:1.4rem; margin-bottom:0.25rem;">AI Candidate Fit Analysis</h2>
      <p style="color:var(--text-muted); font-size:0.9rem; margin-bottom:1.5rem;">Evaluating fit for <strong>${job.title}</strong> at ${job.company_name}</p>
      <div class="form-group"><label class="form-label">Skills</label><input type="text" id="preview-skills" class="form-control" value="${state.user?.skills || "Go, Kubernetes, Docker, PostgreSQL, Microservices"}" /></div>
      <div class="form-group"><label class="form-label">Resume Summary</label><textarea id="preview-resume" class="form-control" rows="4">${state.user?.resume_text || "Senior Backend Engineer with 6+ years in Go, distributed microservices, Docker, Kubernetes, PostgreSQL."}</textarea></div>
      <button class="btn btn-ai" style="width:100%;" id="btn-run-ai" onclick="app.runAIMatchPreview(${jobID})">⚡ Run Instant AI Match Analysis</button>
      <div id="ai-result-box" style="margin-top:1.5rem;"></div>
    `;
    document.getElementById("modal-container").classList.remove("hidden");
  },
  async runAIMatchPreview(jobID) {
    const btn = document.getElementById("btn-run-ai");
    const box = document.getElementById("ai-result-box");
    btn.disabled = true; btn.innerHTML = "Analyzing with AI Model... ⏳";
    try {
      const res = await api.req("/api/jobs/" + jobID + "/ai-match-preview", {
        method: "POST",
        body: JSON.stringify({ candidate_skills: document.getElementById("preview-skills").value, resume_text: document.getElementById("preview-resume").value })
      });
      const scoreClass = res.match_score >= 80 ? "score-high" : (res.match_score >= 60 ? "score-medium" : "score-low");
      box.innerHTML = `
        <div style="background:rgba(255,255,255,0.03); border:1px solid var(--border-color); border-radius:var(--radius-md); padding:1.5rem;">
          <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:1rem;">
            <div class="match-score-badge ${scoreClass}">⚡ ${res.match_score}% Fit Score</div>
            <span style="font-weight:700; color:var(--text-muted);">${res.recommendation}</span>
          </div>
          <p style="font-size:0.92rem; margin-bottom:1rem; color:#e2e8f0;">${res.summary_analysis}</p>
          <div style="margin-bottom:0.75rem;"><div style="font-size:0.8rem; font-weight:700; color:#34d399; text-transform:uppercase;">Matched Strengths:</div>
            <div class="tag-list" style="margin:0.3rem 0;">${res.strengths.map(s => `<span class="tag" style="color:#34d399; border-color:rgba(52,211,153,0.3);">${s}</span>`).join("")}</div></div>
          ${res.missing_skills.length ? `<div style="margin-bottom:1rem;"><div style="font-size:0.8rem; font-weight:700; color:#f87171; text-transform:uppercase;">Identified Skill Gaps:</div><div class="tag-list" style="margin:0.3rem 0;">${res.missing_skills.map(s => `<span class="tag" style="color:#f87171; border-color:rgba(248,113,113,0.3);">${s}</span>`).join("")}</div></div>` : ""}
          <div><div style="font-size:0.8rem; font-weight:700; color:#a5b4fc; text-transform:uppercase; margin-bottom:0.4rem;">AI Screening Questions:</div>
            <ul style="padding-left:1.2rem; font-size:0.85rem; color:var(--text-muted);">${res.interview_questions.map(q => `<li style="margin-bottom:0.3rem;">${q}</li>`).join("")}</ul></div>
        </div>
      `;
    } catch (err) { this.toast(err.message, "error"); }
    finally { btn.disabled = false; btn.innerHTML = "⚡ Run Instant AI Match Analysis"; }
  },
  openApplyModal(jobID) {
    if (!state.user) { this.showAuthModal("login", "Please sign in to apply for this job."); return; }
    if (state.user.role !== "candidate") { this.toast("Only candidates can submit applications.", "error"); return; }
    const job = state.jobs.find(j => j.id === jobID);
    const modal = document.getElementById("modal-content");
    modal.innerHTML = `
      <button class="modal-close" onclick="app.closeModal()">&times;</button>
      <h2 style="font-size:1.4rem; margin-bottom:0.25rem;">Apply for ${job.title}</h2><p style="color:var(--text-muted); font-size:0.9rem; margin-bottom:1.5rem;">${job.company_name} • ${job.location}</p>
      <form onsubmit="app.submitApplication(event, ${jobID})">
        <div class="form-group"><label class="form-label">Core Skills</label><input type="text" id="apply-skills" class="form-control" required value="${state.user.skills || "Go, Kubernetes, Docker, PostgreSQL"}" /></div>
        <div class="form-group"><label class="form-label">Resume / Experience Summary</label><textarea id="apply-resume" class="form-control" rows="5" required>${state.user.resume_text || "Senior Go Engineer with 6+ years in cloud microservices and distributed databases."}</textarea></div>
        <div class="form-group"><label class="form-label">Cover Letter</label><textarea id="apply-cover" class="form-control" rows="3">Excited about building resilient Go architectures with your team.</textarea></div>
        <button type="submit" class="btn btn-primary" style="width:100%;">⚡ Submit Application with AI Fit Evaluation</button>
      </form>
    `;
    document.getElementById("modal-container").classList.remove("hidden");
  },
  async submitApplication(e, jobID) {
    e.preventDefault();
    try {
      const res = await api.req("/api/jobs/" + jobID + "/apply", {
        method: "POST",
        body: JSON.stringify({ candidate_skills: document.getElementById("apply-skills").value, resume_summary: document.getElementById("apply-resume").value, cover_letter: document.getElementById("apply-cover").value })
      });
      this.closeModal(); this.toast("Application submitted! AI Match Score: " + res.application.ai_match_score + "%", "success"); this.navigate("my-apps");
    } catch (err) { this.toast(err.message, "error"); }
  },
  renderPostJob(c) {
    if (!state.user || state.user.role !== "employer") {
      c.innerHTML = `<div style="text-align:center; padding:5rem 1rem;"><h2>Employer Access Required</h2><p style="color:var(--text-muted); margin:1rem 0 1.5rem;">Sign in as an Employer to post jobs and recruit candidates.</p><button class="btn btn-primary" onclick="app.showAuthModal('login')">Sign In</button></div>`;
      return;
    }
    c.innerHTML = `
      <div style="max-width:800px; margin:0 auto; background:var(--bg-card); border:1px solid var(--border-color); border-radius:var(--radius-lg); padding:2.5rem;">
        <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:1.5rem;">
          <div><h2 style="font-size:1.8rem; font-weight:800;">Create Job Posting</h2><p style="color:var(--text-muted);">Use the AI generator or draft specifications manually.</p></div>
          <button class="btn btn-ai btn-sm" onclick="app.openAIGeneratorModal()">✨ AI JD Writer</button>
        </div>
        <form onsubmit="app.submitNewJob(event)">
          <div style="display:grid; grid-template-columns:1fr 1fr; gap:1.2rem;">
            <div class="form-group"><label class="form-label">Job Title *</label><input type="text" id="job-title" class="form-control" required placeholder="e.g. Senior Go Cloud Engineer" /></div>
            <div class="form-group"><label class="form-label">Department</label><input type="text" id="job-dept" class="form-control" placeholder="e.g. Platform Core" /></div>
          </div>
          <div style="display:grid; grid-template-columns:1fr 1fr 1fr; gap:1.2rem;">
            <div class="form-group"><label class="form-label">Location *</label><input type="text" id="job-loc" class="form-control" required placeholder="e.g. Remote / Seattle, WA" /></div>
            <div class="form-group"><label class="form-label">Workplace *</label><select id="job-workplace" class="form-control"><option value="Remote">Remote</option><option value="Hybrid">Hybrid</option><option value="On-site">On-site</option></select></div>
            <div class="form-group"><label class="form-label">Experience *</label><select id="job-exp" class="form-control"><option value="Junior">Junior</option><option value="Mid-Senior">Mid-Senior</option><option value="Senior">Senior</option><option value="Lead">Lead</option></select></div>
          </div>
          <div style="display:grid; grid-template-columns:1fr 1fr; gap:1.2rem;">
            <div class="form-group"><label class="form-label">Salary Min ($)</label><input type="number" id="job-sal-min" class="form-control" value="140000" /></div>
            <div class="form-group"><label class="form-label">Salary Max ($)</label><input type="number" id="job-sal-max" class="form-control" value="195000" /></div>
          </div>
          <div class="form-group"><label class="form-label">Required Skills (Comma separated) *</label><input type="text" id="job-skills" class="form-control" required placeholder="e.g. Go, Kubernetes, Docker, Microservices, PostgreSQL" /></div>
          <div class="form-group"><label class="form-label">Job Description *</label><textarea id="job-desc" class="form-control" rows="5" required placeholder="Mission and responsibilities..."></textarea></div>
          <div class="form-group"><label class="form-label">Requirements *</label><textarea id="job-reqs" class="form-control" rows="4" required placeholder="Qualifications..."></textarea></div>
          <button type="submit" class="btn btn-primary" style="width:100%; padding:0.85rem;">Publish Job Posting</button>
        </form>
      </div>
    `;
  },
  openAIGeneratorModal() {
    const modal = document.getElementById("modal-content");
    modal.innerHTML = `
      <button class="modal-close" onclick="app.closeModal()">&times;</button>
      <h2 style="font-size:1.4rem; margin-bottom:0.25rem;">AI Job Description Generator</h2>
      <p style="color:var(--text-muted); font-size:0.9rem; margin-bottom:1.5rem;">Generate structured job descriptions instantly.</p>
      <div class="form-group"><label class="form-label">Role Title</label><input type="text" id="ai-role-title" class="form-control" placeholder="e.g. Staff Go Platform Architect" /></div>
      <div class="form-group"><label class="form-label">Domain</label><input type="text" id="ai-industry" class="form-control" placeholder="e.g. Cloud Infrastructure" /></div>
      <div class="form-group"><label class="form-label">Key Core Skills</label><input type="text" id="ai-key-skills" class="form-control" placeholder="e.g. Go, Kubernetes, gRPC, Redis, Kafka" /></div>
      <button class="btn btn-ai" style="width:100%;" id="btn-gen-ai" onclick="app.generateAIJD()">✨ Generate Description Now</button>
    `;
    document.getElementById("modal-container").classList.remove("hidden");
  },
  async generateAIJD() {
    const title = document.getElementById("ai-role-title").value;
    const industry = document.getElementById("ai-industry").value;
    const skills = document.getElementById("ai-key-skills").value;
    const btn = document.getElementById("btn-gen-ai");
    btn.disabled = true; btn.innerHTML = "Writing with AI... ⏳";
    try {
      const res = await api.req("/api/ai/generate-jd", {
        method: "POST",
        body: JSON.stringify({ role_title: title, industry: industry, key_skills: skills, experience_level: "Mid-Senior", workplace_type: "Remote" })
      });
      this.closeModal();
      document.getElementById("job-title").value = res.title || title;
      document.getElementById("job-skills").value = res.required_skills || skills;
      document.getElementById("job-desc").value = res.description;
      document.getElementById("job-reqs").value = res.requirements;
      this.toast("AI Job Description generated!", "success");
    } catch (err) { this.toast(err.message, "error"); }
    finally { btn.disabled = false; btn.innerHTML = "✨ Generate Description Now"; }
  },
  async submitNewJob(e) {
    e.preventDefault();
    const payload = {
      title: document.getElementById("job-title").value, department: document.getElementById("job-dept").value,
      location: document.getElementById("job-loc").value, workplace_type: document.getElementById("job-workplace").value,
      employment_type: "Full-time", experience_level: document.getElementById("job-exp").value,
      salary_min: parseInt(document.getElementById("job-sal-min").value) || 0, salary_max: parseInt(document.getElementById("job-sal-max").value) || 0,
      required_skills: document.getElementById("job-skills").value, description: document.getElementById("job-desc").value,
      requirements: document.getElementById("job-reqs").value
    };
    try {
      await api.req("/api/jobs", { method: "POST", body: JSON.stringify(payload) });
      this.toast("Job posted successfully!", "success"); this.navigate("browse");
    } catch (err) { this.toast(err.message, "error"); }
  },
  async renderRecruiterDashboard(c) {
    if (!state.user || state.user.role !== "employer") {
      c.innerHTML = `<div style="text-align:center; padding:5rem 1rem;"><h2>Employer Access Required</h2><p style="color:var(--text-muted); margin:1rem 0 1.5rem;">Sign in as an Employer to view applicants sorted by AI Match Score.</p><button class="btn btn-primary" onclick="app.showAuthModal('login')">Sign In</button></div>`;
      return;
    }
    c.innerHTML = `
      <div class="dashboard-shell">
        <div class="dashboard-header">
          <div>
            <div class="dashboard-kicker">Hiring intelligence</div>
            <h2>Recruiter Pipeline Dashboard</h2>
          </div>
          <div class="dashboard-picker-wrap">
            <label class="form-label">Select Active Job</label>
            <select id="recruiter-job-picker" class="form-control dashboard-select" onchange="app.loadJobApplicants(this.value)">
              <option value="">Select a job...</option>
              ${state.jobs.map(j => `<option value="${j.id}">${j.title} • ${j.company_name}</option>`).join("")}
            </select>
          </div>
        </div>
        <div id="applicant-pipeline-container">
          <div class="empty-state">Choose a job to view applicants.</div>
        </div>
      </div>
    `;
    if (state.jobs.length) {
      document.getElementById("recruiter-job-picker").value = state.jobs[0].id;
      this.loadJobApplicants(state.jobs[0].id);
    }
  },
  async loadJobApplicants(jobID) {
    if (!jobID) return;
    const box = document.getElementById("applicant-pipeline-container");
    box.innerHTML = `<div class="empty-state">Loading applicants...</div>`;
    try {
      const apps = await api.req("/api/jobs/" + jobID + "/applicants");
      if (!apps || !apps.length) {
        box.innerHTML = `<div class="empty-state large">No applications received yet for this role.</div>`;
        return;
      }

      const summary = {
        total: apps.length,
        strong: apps.filter(a => a.ai_match_score >= 80).length,
        reviewing: apps.filter(a => ["reviewing", "shortlisted", "interviewed", "offered"].includes(a.status)).length,
        interview: apps.filter(a => a.status === "interviewed").length
      };

      const pipelineOrder = [
        { key: "applied", label: "Applied", accent: "blue" },
        { key: "reviewing", label: "Reviewing", accent: "violet" },
        { key: "shortlisted", label: "Shortlisted", accent: "green" },
        { key: "interviewed", label: "Interview", accent: "amber" },
        { key: "offered", label: "Offer", accent: "pink" },
        { key: "rejected", label: "Rejected", accent: "red" }
      ];

      const grouped = Object.fromEntries(pipelineOrder.map(p => [p.key, []]));
      apps.forEach(a => {
        const key = a.status && grouped[a.status] !== undefined ? a.status : "applied";
        grouped[key].push(a);
      });

      box.innerHTML = `
        <div class="dashboard-metrics">
          <div class="metric-card">
            <span>Total Applicants</span>
            <strong>${summary.total}</strong>
            <small>Across this role</small>
          </div>
          <div class="metric-card accent-cyan">
            <span>Strong Fit</span>
            <strong>${summary.strong}</strong>
            <small>80%+ AI match</small>
          </div>
          <div class="metric-card accent-violet">
            <span>In Review</span>
            <strong>${summary.reviewing}</strong>
            <small>Active pipeline</small>
          </div>
          <div class="metric-card accent-amber">
            <span>Interview Stage</span>
            <strong>${summary.interview}</strong>
            <small>Scheduled interviews</small>
          </div>
        </div>
        <div class="pipeline-board">
          ${pipelineOrder.map(stage => {
            const cards = grouped[stage.key] || [];
            return `
              <div class="pipeline-column ${stage.accent}">
                <div class="pipeline-head">
                  <span>${stage.label}</span>
                  <strong>${cards.length}</strong>
                </div>
                <div class="pipeline-stack">
                  ${cards.length ? cards.map(a => {
                    const scoreClass = a.ai_match_score >= 80 ? "score-high" : (a.ai_match_score >= 60 ? "score-medium" : "score-low");
                    let strengths = []; try { strengths = JSON.parse(a.ai_strengths); } catch (e) {}
                    return `
                      <div class="candidate-card">
                        <div class="candidate-topline">
                          <div>
                            <h3>${a.candidate_name}</h3>
                            <p>${a.candidate_email}</p>
                          </div>
                          <span class="match-score-badge ${scoreClass}">${a.ai_match_score}%</span>
                        </div>
                        <div class="candidate-meta">${a.candidate_skills || "Skills not specified"}</div>
                        <div class="tag-list compact">${strengths.slice(0, 3).map(s => `<span class="tag">${s}</span>`).join("") || '<span class="tag">AI profile ready</span>'}</div>
                        <div class="candidate-actions">
                          <select class="form-control compact-select" onchange="app.changeAppStatus(${a.id}, this.value)">
                            <option value="applied" ${a.status === "applied" ? "selected" : ""}>Applied</option>
                            <option value="reviewing" ${a.status === "reviewing" ? "selected" : ""}>Reviewing</option>
                            <option value="shortlisted" ${a.status === "shortlisted" ? "selected" : ""}>Shortlisted</option>
                            <option value="interviewed" ${a.status === "interviewed" ? "selected" : ""}>Interview Scheduled</option>
                            <option value="offered" ${a.status === "offered" ? "selected" : ""}>Offer Extended</option>
                            <option value="rejected" ${a.status === "rejected" ? "selected" : ""}>Rejected</option>
                          </select>
                          <button class="btn btn-outline btn-sm" onclick='app.viewCandidateDetail(${JSON.stringify(a)})'>Profile</button>
                        </div>
                      </div>
                    `;
                  }).join("") : `<div class="empty-mini">No candidates</div>`}
                </div>
              </div>
            `;
          }).join("")}
        </div>
      `;
    } catch (err) { this.toast(err.message, "error"); }
  },
  async changeAppStatus(appID, status) {
    try {
      await api.req("/api/applications/" + appID + "/status", { method: "PUT", body: JSON.stringify({ status: status, notes: "Status updated to " + status }) });
      this.toast("Candidate status updated to " + status, "success");
    } catch (err) { this.toast(err.message, "error"); }
  },
  viewCandidateDetail(a) {
    const modal = document.getElementById("modal-content");
    modal.innerHTML = `
      <button class="modal-close" onclick="app.closeModal()">&times;</button>
      <h2 style="font-size:1.4rem; margin-bottom:0.25rem;">${a.candidate_name}</h2><p style="color:var(--text-muted); font-size:0.9rem; margin-bottom:1rem;">${a.candidate_email}</p>
      <div style="margin-bottom:1rem;"><div style="font-size:0.8rem; font-weight:700; color:var(--text-muted); text-transform:uppercase;">AI Evaluation Summary</div><p style="background:rgba(255,255,255,0.03); border:1px solid var(--border-color); padding:1rem; border-radius:var(--radius-md); font-size:0.9rem; margin-top:0.4rem;">${a.notes || "Strong candidate alignment."}</p></div>
      <div style="margin-bottom:1rem;"><div style="font-size:0.8rem; font-weight:700; color:var(--text-muted); text-transform:uppercase;">Resume Experience</div><div style="background:rgba(255,255,255,0.03); border:1px solid var(--border-color); padding:1rem; border-radius:var(--radius-md); font-size:0.9rem; margin-top:0.4rem; white-space:pre-wrap;">${a.resume_summary || "No resume text provided."}</div></div>
      ${a.cover_letter ? `<div><div style="font-size:0.8rem; font-weight:700; color:var(--text-muted); text-transform:uppercase;">Cover Letter</div><p style="background:rgba(255,255,255,0.03); border:1px solid var(--border-color); padding:1rem; border-radius:var(--radius-md); font-size:0.9rem; margin-top:0.4rem;">${a.cover_letter}</p></div>` : ""}
    `;
    document.getElementById("modal-container").classList.remove("hidden");
  },
  async renderCandidateApps(c) {
    if (!state.user) { this.showAuthModal("login", "Sign in to track your job applications."); return; }
    c.innerHTML = `<div style="margin-bottom:2rem;"><h2 style="font-size:2rem; font-weight:800;">My Applications</h2><p style="color:var(--text-muted);">Track your status and AI evaluation reports.</p></div><div id="candidate-apps-list"><div style="text-align:center; padding:3rem; color:var(--text-muted);">Loading applications...</div></div>`;
    try {
      const apps = await api.req("/api/candidate/applications");
      const box = document.getElementById("candidate-apps-list");
      if (!apps || !apps.length) { box.innerHTML = `<div style="text-align:center; padding:3rem; color:var(--text-muted); background:var(--bg-card); border-radius:var(--radius-md);">You have not applied to any roles yet. Explore jobs to get started!</div>`; return; }
      box.innerHTML = `<div style="display:flex; flex-direction:column; gap:1.2rem;">${apps.map(a => `
        <div style="background:var(--bg-card); border:1px solid var(--border-color); border-radius:var(--radius-lg); padding:1.5rem; display:flex; justify-content:space-between; align-items:center;">
          <div><div style="font-size:0.85rem; font-weight:700; color:var(--accent-secondary);">${a.company_name}</div><h3 style="font-size:1.2rem; font-weight:700; margin:0.2rem 0 0.5rem;">${a.job_title}</h3><p style="color:var(--text-muted); font-size:0.88rem;">${a.notes || "Application submitted to hiring system."}</p></div>
          <div style="text-align:right;"><div class="match-score-badge score-high" style="font-size:0.9rem; padding:0.35rem 0.85rem; margin-bottom:0.5rem;">⚡ ${a.ai_match_score}% Fit</div><div><span class="tag tag-accent">${a.status.toUpperCase()}</span></div></div>
        </div>
      `).join("")}</div>`;
    } catch (err) { this.toast(err.message, "error"); }
  },
  showAuthModal(mode = "login", msg = "") {
    const modal = document.getElementById("modal-content");
    modal.innerHTML = `
      <button class="modal-close" onclick="app.closeModal()">&times;</button>
      <h2 style="font-size:1.5rem; margin-bottom:0.25rem;">${mode === "login" ? "Welcome Back" : "Create an Account"}</h2>
      <p style="color:var(--text-muted); font-size:0.9rem; margin-bottom:1.5rem;">${msg || (mode === "login" ? "Sign in to access your jobs and applications" : "Join TalentPulse AI today")}</p>
      <div style="display:flex; gap:0.5rem; margin-bottom:1.5rem; background:rgba(255,255,255,0.05); padding:0.3rem; border-radius:var(--radius-md);">
        <button class="btn btn-sm ${mode === "login" ? "btn-primary" : "btn-outline"}" style="flex:1;" onclick="app.showAuthModal('login')">Sign In</button>
        <button class="btn btn-sm ${mode === "register" ? "btn-primary" : "btn-outline"}" style="flex:1;" onclick="app.showAuthModal('register')">Register</button>
      </div>
      <form onsubmit="app.handleAuthSubmit(event, '${mode}')">
        ${mode === "register" ? `
          <div class="form-group"><label class="form-label">Full Name</label><input type="text" id="auth-name" class="form-control" required placeholder="Jane Doe" /></div>
          <div class="form-group"><label class="form-label">I am a</label><select id="auth-role" class="form-control" onchange="app.toggleRoleFields(this.value)"><option value="candidate">Job Seeker (Candidate)</option><option value="employer">Recruiter / Employer</option></select></div>
          <div id="company-field" class="form-group" style="display:none;"><label class="form-label">Company Name</label><input type="text" id="auth-company" class="form-control" placeholder="Acme Tech Inc" /></div>
        ` : ""}
        <div class="form-group"><label class="form-label">Email Address</label><input type="email" id="auth-email" class="form-control" required placeholder="you@domain.com" /></div>
        <div class="form-group"><label class="form-label">Password</label><input type="password" id="auth-password" class="form-control" required placeholder="••••••••" /></div>
        <button type="submit" class="btn btn-primary" style="width:100%; margin-top:0.5rem;">${mode === "login" ? "Sign In" : "Create Account"}</button>
      </form>
      ${mode === "login" ? `<div style="margin-top:1.5rem; padding-top:1rem; border-top:1px solid var(--border-color); font-size:0.85rem; color:var(--text-muted); text-align:center;"><p style="margin-bottom:0.5rem;">Quick Demo Accounts (Password: <code>password123</code>):</p><div style="display:flex; justify-content:center; gap:0.5rem;"><button class="btn btn-outline btn-sm" onclick="app.fillDemo('recruiter@google.com')">Recruiter Demo</button><button class="btn btn-outline btn-sm" onclick="app.fillDemo('alex.dev@gmail.com')">Candidate Demo</button></div></div>` : ""}
    `;
    document.getElementById("modal-container").classList.remove("hidden");
  },
  toggleRoleFields(role) { const comp = document.getElementById("company-field"); if (comp) comp.style.display = role === "employer" ? "block" : "none"; },
  fillDemo(email) { document.getElementById("auth-email").value = email; document.getElementById("auth-password").value = "password123"; },
  async handleAuthSubmit(e, mode) {
    e.preventDefault();
    const email = document.getElementById("auth-email").value; const password = document.getElementById("auth-password").value;
    const body = { email, password };
    if (mode === "register") {
      body.full_name = document.getElementById("auth-name").value;
      body.role = document.getElementById("auth-role").value;
      if (body.role === "employer") body.company_name = document.getElementById("auth-company")?.value || "";
    }
    try {
      const res = await api.req(mode === "login" ? "/api/auth/login" : "/api/auth/register", { method: "POST", body: JSON.stringify(body) });
      state.user = res.user; state.token = res.token; localStorage.setItem("tp_user", JSON.stringify(res.user)); localStorage.setItem("tp_token", res.token);
      this.closeModal(); this.renderNavAuth(); this.toast("Welcome, " + res.user.full_name + "!", "success"); this.navigate("browse");
    } catch (err) { this.toast(err.message, "error"); }
  }
};
window.app = app;
document.addEventListener("DOMContentLoaded", () => app.init());
