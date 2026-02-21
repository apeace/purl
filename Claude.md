# Claude Guidelines

## Architecture

- This is an NPM workspaces monorepo with multiple frontend apps. Shared code belongs in workspace packages, not duplicated across apps.
- Extract shared UI and logic into reusable components. Before building something new, check whether a reusable component already exists.
- Before working in any app directory, read its CLAUDE.md if one exists — it contains app-specific rules that take precedence over these.

## UI

- All UI must be polished and visually refined. Add subtle, purposeful animations (transitions, hover states, etc.) to interactive elements.
- Design mobile-first. Size elements and layout to avoid unnecessary scrolling on small screens.
- Desktop is equally important — always consider both viewports. Use the extra space on desktop to improve the experience, not just scale things up.
- Account for platform differences: hover interactions are desktop-only. Use hover to enhance the desktop experience, but never rely on it to convey functionality that mobile users also need.

## Code Style

- Use double quotes and omit semicolons everywhere.
- Alphabetize imports. Remove unused imports immediately — never leave them in.
- Only add a dependency when it provides enough value to justify the cost. Prefer a small amount of hand-written code over pulling in a new package.
- Comment non-obvious logic, but not self-explanatory code. Over-commenting is as bad as under-commenting.
- When modifying code, verify that nearby existing comments are still accurate. Update or remove any that are stale.
- Never write functions or code that isn't called. Every piece of code must serve a current, concrete purpose.

## Git

- After creating any new file, immediately run `git add <file>` to stage it.
