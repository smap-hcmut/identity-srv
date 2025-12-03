package collector

import "time"

// Post represents a social media post collected by the Collector Service.
type Post struct {
	ID         string     `json:"id"`
	Content    string     `json:"content"`
	Source     string     `json:"source"`
	SourceID   string     `json:"source_id"`
	Author     string     `json:"author"`
	AuthorID   string     `json:"author_id"`
	Date       time.Time  `json:"date"`
	URL        string     `json:"url"`
	Engagement Engagement `json:"engagement"`
	Metadata   Metadata   `json:"metadata"`
}

// Engagement represents the engagement metrics of a post.
type Engagement struct {
	Likes    int `json:"likes"`
	Comments int `json:"comments"`
	Shares   int `json:"shares"`
}

// Metadata represents additional metadata of a post.
type Metadata struct {
	Language  string `json:"language"`
	Sentiment string `json:"sentiment"`
}

// DryRunRequest is the request body for the dry run endpoint.
type DryRunRequest struct {
	Keywords []string `json:"keywords"`
	Limit    int      `json:"limit"`
}

// DryRunResponse is the response body for the dry run endpoint.
type DryRunResponse struct {
	Posts      []Post `json:"posts"`
	TotalFound int    `json:"total_found"`
	Limit      int    `json:"limit"`
}
