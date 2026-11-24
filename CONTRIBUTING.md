# Contributing to Tapr

Thank you for showing interest in contributing to **Tapr**! 🎉  
Whether it’s fixing a bug, adding a new feature, improving documentation, or helping with tests, your contributions are **welcome and appreciated**.

---

## Why Contribute?

- ✨ Work on a **fast-growing CLI tool** for APIs and DevOps  
- 🌍 Join a community of developers passionate about **observability and monitoring**  
- 🏆 Get recognition for your contributions on GitHub  
- 🎁 Help shape Tapr into the **go-to developer CLI tool**  

No contribution is too small- even fixing a typo counts!

---

## How to Get Started

**Prerequisites:** Go 1.21+ and git.

1. **Find an issue.** Look for [`good first issue`](https://github.com/symtalha14/tapr/labels/good%20first%20issue) or [`help wanted`](https://github.com/symtalha14/tapr/labels/help%20wanted). Comment on it so others know you're working on it.
2. **Fork and clone**

   ```bash
   git clone https://github.com/<your-username>/tapr.git
   cd tapr
   ```

3. **Create a branch**

   ```bash
   git checkout -b fix/short-description
   ```

4. **Build and run locally**

   ```bash
   go build -o tapr ./cmd/tapr/
   ./tapr https://httpbin.org/get
   ./tapr batch sample/endpoints.yaml
   ```

5. **Run the checks** (the same ones CI runs)

   ```bash
   gofmt -l .            # should print nothing; fix with: gofmt -w .
   go vet ./...
   go test -race ./...
   ```

   Prefer `net/http/httptest` in tests over hitting real endpoints.

6. **Commit** using [Conventional Commits](https://www.conventionalcommits.org/) prefixes (`feat:`, `fix:`, `docs:`, `test:`, `chore:`):

   ```bash
   git commit -m "fix: clamp box padding for long URLs"
   ```

7. **Push and open a Pull Request** against `main`. Link the issue (`Fixes #123`) and fill in the PR template.

## Coding Standards

- Adhere to Go conventions (gofmt) and idiomatic patterns
- Use meaningful variable names instead of random names, even if they are slightly longer
- Write clear, maintainable, and well-documented code
- Include unit tests for all new functionality
- Use descriptive commit messages that summarize your changes
- Keep the CLI interface intuitive and minimal

## Issue Reporting

When submitting a bug report or feature request (use the issue templates):
- Provide steps to reproduce (if applicable)
- Clearly describe expected vs actual behavior
- Include environment details (OS, Go version, Tapr version)

## Review Process
All contributions are reviewed by the maintainers. Feedback may be provided, and revisions requested, to ensure consistency, quality, and maintainability.


---

# Community Standards / Code of Conduct

All contributors, maintainers, and community members are expected to uphold the following principles:

1. Be Respectful and Professional
2. Treat all participants with respect regardless of experience, background, or opinions.
3. Avoid offensive, harassing, or discriminatory language.
4. Collaborate Constructively
5. Focus feedback on the work, not the individual.
6. Be open to suggestions and maintain a collaborative attitude.
7. Be Inclusive
8. Encourage participation from diverse contributors.
9. Avoid language, jokes, or behavior that could make others feel unwelcome.
10. Maintain Professional Communication
11. Use clear, concise, and courteous language in issues, PRs, and discussions.
12. Avoid spamming, trolling, or unnecessary off-topic posts.
13. Follow Project Guidelines
14. Adhere to coding standards, documentation rules, and PR requirements.
15. Respect maintainers’ decisions and project workflow.
16. Report Violations Responsibly
17. Keep reports professional, factual, and confidential.

Help foster a positive and collaborative community and let's build Taper together. 💪
