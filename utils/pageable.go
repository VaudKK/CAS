package utils

type Pageable struct {
	Page   int
	Size   int
	OffSet int
}

type PageInfo struct {
	CurrentPage int `json:"currentPage"`
	Size        int `json:"size"`
	FirstPage   int `json:"firstPage"`
	LastPage    int `json:"lastPage"`
	TotalItems  int `json:"totalItems"`
}
