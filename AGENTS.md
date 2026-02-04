# AGENTS.md

## Purpose

This repository contains a Go implementation of an HTML5 parser (`h5p`).
The project prioritises spec-driven correctness, deterministic behaviour, and high-quality, readable tests.

This document defines **mandatory rules** for AI agents and human contributors working in this repository.

---

## General Rules

- Do not change public APIs unless explicitly requested.
- Prefer clarity over cleverness.
- Follow idiomatic Go (`gofmt`, standard library first).
- New functionality must be fully tested.
- Existing tests must not be weakened, removed, or made less precise.

---

## Testing Requirements (MANDATORY)

### Test Structure

All Go tests **must** follow these rules:

- Use `t.Run(...)` with descriptive, human-readable names.
- Use a slice of test cases (table-driven tests).
- Avoid duplicated test logic.
- Fail fast with clear error messages.

---

### Required Test Pattern

```go
func TestSomething(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple valid input",
			input:    "<div></div>",
			expected: "<div></div>",
		},
		{
			name:     "self-closing tag",
			input:    "<br/>",
			expected: "<br>",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // where safe

			result := Parse(tc.input)
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}
```

---

### Do Not

- Write one-off `TestXxx` functions without `t.Run`.
- Use cryptic or non-descriptive test names.
- Hide test intent behind excessive helper abstractions.
- Duplicate logic across multiple test functions.
- Silence failures with broad conditionals.

---

## Descriptive Test Names

Test case names must describe **observable behaviour**, not implementation details.

### Good Examples

- handles malformed end tag
- preserves whitespace in text nodes
- parses deeply nested elements correctly

### Bad Examples

- case 1
- testA
- edge case

---

## Coverage Expectations

- New code must not reduce overall test coverage.
- Every new logical branch must be covered by tests.
- Pull requests that introduce uncovered code paths will be rejected.

When in doubt:
1. Write the test first.
2. Then implement the behaviour.

---

## Spec-Driven Behaviour

Where behaviour is derived from:
- The HTML5 specification
- Established browser behaviour
- Known parsing quirks

Tests must encode the behaviour directly.

---

## Parallelism

- Use `t.Parallel()` only when safe.
- Never share mutable global state across tests.
- Avoid order-dependent tests.

---

## Error Handling Tests

When testing error cases:

- Always assert that an error exists.
- Assert error type or message where meaningful.
- Never ignore returned errors.

---

## Pull Request Discipline

Every pull request must:

- Compile cleanly.
- Pass all tests.
- Maintain or improve test coverage.
- Follow the conventions defined in this document.
- Avoid unrelated refactors or formatting-only changes.

---

## When In Doubt

- Match the existing test style in the repository.
- Prefer explicit code over implicit behaviour.
- Ask for clarification rather than guessing.

---

**This file is authoritative.  
AI agents and contributors are expected to comply strictly.**
