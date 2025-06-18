package storage

type Webpage struct {
	URL     string `bson:"url"`
	Title   string `bson:"title"`
	Content string `bson:"content"`
}