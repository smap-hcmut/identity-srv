# Tasks

- [ ] **Layer 1: AI Suggestion** <!-- id: 0 -->
  - [ ] Implement `KeywordSuggester` service using LLM provider.
  - [ ] Create API endpoint `POST /projects/keywords/suggest`.
- [ ] **Layer 2: Semantic Validator** <!-- id: 1 -->
  - [ ] Update `KeywordValidator` with length and stopword checks.
  - [ ] Integrate LLM check for ambiguity.
  - [ ] Update `POST /projects` to enforce validation.
- [ ] **Layer 3: Dry Run** <!-- id: 2 -->
  - [ ] Implement `KeywordTester` service.
  - [ ] Create API endpoint `POST /projects/keywords/dry-run`.
  - [ ] Mock Collector Service response for dry run (or integrate if available).
- [ ] **Data Model** <!-- id: 3 -->
  - [ ] Add `ExcludeKeywords` to `Project` model and database migration.
  - [ ] Update Create/Update APIs to accept `exclude_keywords`.
