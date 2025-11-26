# Design: Smart Keyword Enforcer

## Overview

The Smart Keyword Enforcer introduces a 3-layer architecture within the **Project Service** to validate and optimize keywords before they are used for data collection.

## Architecture Layers

### Layer 1: AI Suggestion & Expansion (The Helper)

- **Component**: `KeywordSuggester`
- **Function**: Connects to an LLM (e.g., Gemini Flash/GPT-4o-mini).
- **Input**: Brand Name (e.g., "VinFast").
- **Output**:
  - **Niche Keywords**: Specific variations (e.g., "VinFast VF3", "VF Wild").
  - **Negative Keywords**: Terms to exclude (e.g., "sim vinfast", "xổ số").
- **Benefit**: Reduces user effort and improves keyword coverage.

### Layer 2: Semantic Validator (The Gatekeeper)

- **Component**: `KeywordValidator`
- **Function**: Checks for generic terms and potential ambiguities.
- **Rules**:
  - **Length Check**: Warn if too short (1 word).
  - **Stopwords**: Block common words (e.g., "xe", "mua").
  - **LLM Check**: Ask LLM if the keyword is ambiguous (e.g., "Apple" -> fruit vs tech).

### Layer 3: Dry Run (The Reality Check)

- **Component**: `KeywordTester`
- **Function**: Fetches a small sample of real data to show the user what the keyword returns.
- **Flow**:
  1. User clicks "Test Keyword".
  2. Project Service calls **Collector Service** with `dry_run=true`.
  3. Collector fetches 5-10 recent posts.
  4. UI displays posts.
  5. User verifies relevance.

## User Flow Update

1.  **Basic Info**: Name, Date Range.
2.  **Smart Configuration**:
    - User enters Seed Keyword.
    - System suggests Niche & Negative Keywords (Layer 1).
    - User selects/deselects.
3.  **Dry Run**:
    - System fetches sample posts (Layer 3).
    - User reviews. If "garbage", User adds Negative Keywords (Layer 2 warns if still generic).
4.  **Submit**:
    - Final payload includes `Include_Keywords` and `Exclude_Keywords`.

## Data Model Changes

Update `Project` struct to include `ExcludeKeywords`.

```json
{
  "brand_keywords": ["VinFast", "VF3"],
  "exclude_keywords": ["sim", "xổ số"]
}
```
