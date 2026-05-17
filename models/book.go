package book_proposals

type Book struct {
	Title         string `json:"title"`
	Author        string `json:"author"`
	Rating        int    `json:"rating"`
	Status        string `json:"status"`
	YearPublished string `json:"year_published"`
	Shelves       string `json:"shelves"`
}
