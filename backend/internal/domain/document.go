package domain

type Document struct {
	ID              int64
	UserID          int64
	Filename        string
	NullEmbeddings  int64
	TotalEmbeddings int64
}
