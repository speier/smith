# Smith -- Product Ideas

## Top Pick: AI DevOps Agent

**"Your $10/mo on-call engineer"**

### Problem

Solo devs, small teams, and indie hackers deploy apps but have NO monitoring, no on-call, no one watching their stuff at 3am. Downtime costs them customers and money. Existing monitoring tools (Datadog, PagerDuty) are expensive, complex, and built for large teams.

### What It Does

- Monitors deployed apps (HTTP health checks, SSL cert expiry, DNS, error rates)
- When something breaks, Smith **diagnoses the issue** using AI (reads logs, checks recent deploys, analyzes error patterns)
- Sends a notification with **what's wrong AND a suggested fix** (not just "your site is down")
- Can optionally auto-fix common issues (restart service, rollback deploy, clear cache)
- Learns your infrastructure patterns over time

### Why It Works

- Clear, acute pain point -- downtime = lost money
- $10/mo is nothing compared to cost of even 1 hour of downtime
- Reuses existing agent orchestration, tool execution, and safety system
- Low competition in "AI-powered monitoring for indie devs" niche
- Extremely sticky -- once set up, nobody cancels
- Brand fits perfectly: "Smith is watching your servers"

### Pricing

| Tier | Price | Includes |
|------|-------|---------|
| Free | $0 | 1 monitor, basic health checks |
| Pro | $10/mo | 10 monitors, AI diagnosis, notifications |
| Team | $20/mo | Unlimited monitors, auto-remediation, Slack/Discord integration |

### Reuses From Current Codebase

- Agent orchestration (multi-agent coordination)
- Tool execution framework with safety levels
- BBolt storage for state/history
- Event bus for real-time notifications
- LLM provider abstraction (diagnosis powered by any model)

---

## Other Ideas

### AI Codebase Onboarding

**"Understand any codebase in 5 minutes"**

Point Smith at any repo. It analyzes the entire codebase and builds an interactive knowledge base. Ask questions in natural language, get architecture diagrams, dependency maps, and contributor guides. Stays updated as code changes.

- Universal dev pain point
- Free (small repos) / $5/mo (unlimited) / $20/seat/mo (teams)
- Minimal infrastructure -- runs locally, calls LLM APIs
- Open source maintainers would love this for onboarding contributors

### AI PR Reviewer

**"The teammate who actually reads your PRs"**

GitHub App that automatically reviews every PR with deep codebase context. Catches bugs, security issues, performance problems. Learns your team's conventions and enforces them. Zero setup.

- GitHub App = viral distribution via marketplace
- Free (public repos) / $10/mo/repo (private) / $20/mo unlimited
- High competition but high willingness to pay

### Personal Automation Agent

**"Your AI that actually DOES things on your computer"**

Runs locally. Describe tasks in plain English. Smith creates and executes automated workflows using local tools. Safety system asks before doing anything dangerous. Privacy-first (runs on your machine).

- Broad market beyond developers
- $10/mo for hours saved per week
- Existing safety system and tool execution are a natural fit

### AI Changelog & Release Notes

**"Never write release notes again"**

Connects to Git repo, auto-generates user-facing changelogs from commits/PRs. Categorizes changes, generates marketing-friendly notes, posts to Slack/Discord/email. CLI tool + GitHub Action.

- Tiny focused problem, very solvable
- $5/mo impulse buy
- Every SaaS company needs this
- Fastest MVP (2-3 weeks)

---

## Comparison Matrix

| Criteria | DevOps Agent | Codebase Onboarding | PR Reviewer | Personal Auto | Changelog |
|----------|-------------|---------------------|-------------|---------------|-----------|
| Pain severity | 9/10 | 8/10 | 7/10 | 6/10 | 5/10 |
| Willingness to pay | 9/10 | 7/10 | 8/10 | 6/10 | 7/10 |
| Competition | Low | Medium | High | Medium | Medium |
| Reuses existing code | High | High | Medium | High | Low |
| Time to MVP | 4-6 weeks | 3-4 weeks | 6-8 weeks | 4-6 weeks | 2-3 weeks |
| Stickiness | Very High | High | High | Medium | Medium |
