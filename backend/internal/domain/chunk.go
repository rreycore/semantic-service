package domain

type Chunk struct {
	ID         int64
	UserID     int64
	DocumentID int64
	Text       string
}

type SearchResult struct {
	ID         int64
	DocumentID int64
	Title      string
	Text       string
	Distance   float64
}
