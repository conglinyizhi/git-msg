---
name: clyzhi-git
description: Git Commit Message Specification (Lightweight Modification Based on Conventional Commits)
metadata:
  author: conglinyizhi
  version: 2.0.1
  created: 2026-02-13
  updated: 2026-02-13
---

# Git Commit Message Specification

Follow the **Conventional Commits** format. Unless there are special requirements, git commit messages should be submitted according to this scheme.  
Additionally, avoid invoking this capability (or attempting to influence the git repository in any way) without explicit user instructions.

## Basic Format

`<type>[scope]: <subject>`

## Required Fields

### Type

The type must use the categories mentioned in the specification. Commonly used types include:

| Type     | Purpose                 | Description                                                                                                     |
| -------- | ----------------------- | --------------------------------------------------------------------------------------------------------------- |
| fix      | Bug Fix                 | Fixes a bug in the codebase                                                                                     |
| feat     | New Feature             | Adds a new feature to the codebase (a user-visible new feature)                                                 |
| build    | Build System            | Changes affecting the build system or external dependencies (e.g., webpack, gulp, npm)                          |
| chore    | Chores                  | Changes that do not modify source or test files, such as updating dependencies or auxiliary tool configurations |
| ci       | Continuous Integration  | Changes to CI configuration files and scripts (e.g., Travis, Jenkins, Circle)                                   |
| docs     | Documentation           | Documentation-only changes, such as README, API documentation, code comments                                    |
| style    | Code Style              | Changes that do not affect the meaning of the code, such as whitespace, formatting, or missing semicolons       |
| refactor | Refactor                | Code changes that neither fix a bug nor add a feature                                                           |
| perf     | Performance Improvement | Code changes that improve performance                                                                           |
| test     | Tests                   | Adding or correcting test cases                                                                                 |
| revert   | Revert                  | Reverts a previous commit                                                                                       |

### Subject

- Use English for descriptions
- Keep it concise and clear, no more than 50 characters
- Do not end with a period
- Use imperative mood

## Optional Fields

### Scope

Scope is used to indicate the specific module, component, or location affected by the commit. It is optional. When the change involves a specific scope, it is recommended to include it to provide clearer information.

**Scope Naming Rules:**

- Use lowercase letters, numbers, and hyphens (-)
- Keep it concise and accurately reflect the affected scope
- Multiple scopes can be separated by commas (e.g., feat(auth,api):)

**Common Scope Examples:**

| Scope Example | Description                                          |
| ------------- | ---------------------------------------------------- |
| api           | Related to API interfaces                            |
| ui            | Related to the user interface                        |
| auth          | Related to authentication and authorization          |
| db            | Related to the database                              |
| config        | Related to configuration files                       |
| deps          | Related to dependencies                              |
| docs          | Related to documentation (when the type is not docs) |
| utils         | Related to utility functions                         |
| types         | Related to TypeScript type definitions               |
| components    | Related to components (frontend)                     |
| hooks         | Related to custom hooks                              |
| store         | Related to state management                          |
| styles        | Related to styles                                    |
| middleware    | Related to middleware                                |
| routes        | Related to routing                                   |
| services      | Related to the service layer                         |
| models        | Related to data models                               |
| migrations    | Related to database migrations                       |
| tests         | Related to tests (when the type is not test)         |
| workflows     | Related to CI/CD workflows                           |
| docker        | Related to Docker                                    |
| scripts       | Related to scripts                                   |

**Usage Examples:**

- `feat(auth): Add mobile number login functionality`
- `fix(api): Fix incorrect data format in user list API response`
- `style(components): Unify spacing for button components`
- `docs(README): Update installation instructions`

# Additional Output Notes

**When outputting, only include the information to be placed in `git commit -m "<message>"`. Avoid outputting any other content that might interfere with the workflow.**
