# Generation Prompt

**Model:** Gemini 3.1 Pro (High)
**Date:**  2026-04-26

## Prompt

You are an expert software engineer specialising in concurrent and distributed systems.

## Task
Translate the academic paper at the URL below into a complete, idiomatic Go implementation.

## Paper
https://drops.dagstuhl.de/storage/00lipics/lipics-vol046-opodis2015/LIPIcs.OPODIS.2015.35/LIPIcs.OPODIS.2015.35.pdf

## Requirements

### Code
- Implement every data structure, algorithm, and operation described in the paper
- Map each pseudocode function to a named Go function with a comment citing the
  corresponding figure and line numbers from the paper
- Use idiomatic Go: sync/atomic for all shared state, unsafe.Pointer for CAS on
  pointer-width fields, no locks unless the paper explicitly uses them
- Export only the public API described in the paper's specification section;
  keep internal helpers unexported
- Add a package-level doc comment summarising the paper (title, authors, venue, DOI)

### Tests
- Write a _test.go file covering:
  - Sequential correctness (all operations, boundary conditions)
  - Concurrent stress tests using t.Parallel() and sync.WaitGroup
  - Race-detector clean: all tests must pass under `go test -race`
  - At least one Example function demonstrating the primary use case

### Structure
Produce exactly three files:
1. `<package>.go`      — implementation
2. `<package>_test.go` — tests
3. `go.mod`            — module file (module path: github.com/example/<package>)

### Traceability
Every non-trivial function must include an inline comment of the form:
  // Paper: Fig. <N>, lines <X>-<Y>
linking it back to the source.

## Constraints
- Do not invent behaviour not present in the paper
- If the paper omits implementation details (e.g. memory reclamation),
  state the assumption explicitly in a code comment
- Flag any deviation from the paper's pseudocode with a // NOTE: comment

## Deliverable
Return the three files and confirm all tests pass under `go test -race ./...`

## Notes
Migrated to new agents.md structure.