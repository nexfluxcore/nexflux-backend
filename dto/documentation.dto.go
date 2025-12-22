package dto

import "time"

// ==================================================
// DOCUMENTATION CATEGORY DTOs
// ==================================================

// DocCategoryResponse - response for documentation category
type DocCategoryResponse struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Description   string `json:"description"`
	Icon          string `json:"icon"`
	Color         string `json:"color"`
	Order         int    `json:"order"`
	ArticlesCount int    `json:"articles_count"`
}

// DocCategoryDetailResponse - category with articles
type DocCategoryDetailResponse struct {
	ID            string                    `json:"id"`
	Name          string                    `json:"name"`
	Slug          string                    `json:"slug"`
	Description   string                    `json:"description"`
	Icon          string                    `json:"icon"`
	Color         string                    `json:"color"`
	Articles      []DocArticleBriefResponse `json:"articles"`
	ArticlesCount int                       `json:"articles_count"`
}

// ==================================================
// DOCUMENTATION ARTICLE DTOs
// ==================================================

// DocArticleListRequest - request for listing articles
type DocArticleListRequest struct {
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
	Category   string `form:"category"`
	Difficulty string `form:"difficulty"`
	Search     string `form:"search"`
	Sort       string `form:"sort"` // newest, popular, title
}

// DocArticleBriefResponse - article brief info for lists
type DocArticleBriefResponse struct {
	ID              string                    `json:"id"`
	Title           string                    `json:"title"`
	Slug            string                    `json:"slug"`
	Excerpt         string                    `json:"excerpt"`
	Category        *DocCategoryBriefResponse `json:"category,omitempty"`
	Author          *DocAuthorResponse        `json:"author,omitempty"`
	ReadTimeMinutes int                       `json:"read_time_minutes"`
	Difficulty      string                    `json:"difficulty"`
	Tags            []string                  `json:"tags"`
	Views           int                       `json:"views"`
	IsFeatured      bool                      `json:"is_featured"`
	IsNew           bool                      `json:"is_new,omitempty"`
	PublishedAt     *time.Time                `json:"published_at"`
}

// DocCategoryBriefResponse - minimal category info
type DocCategoryBriefResponse struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// DocAuthorResponse - author info
type DocAuthorResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// DocArticleDetailResponse - full article details
type DocArticleDetailResponse struct {
	ID              string                      `json:"id"`
	Title           string                      `json:"title"`
	Slug            string                      `json:"slug"`
	Excerpt         string                      `json:"excerpt"`
	Content         string                      `json:"content"`
	Category        *DocCategoryBriefResponse   `json:"category"`
	Author          *DocAuthorResponse          `json:"author"`
	ReadTimeMinutes int                         `json:"read_time_minutes"`
	Difficulty      string                      `json:"difficulty"`
	Tags            []string                    `json:"tags"`
	Views           int                         `json:"views"`
	IsFeatured      bool                        `json:"is_featured"`
	PublishedAt     *time.Time                  `json:"published_at"`
	UpdatedAt       time.Time                   `json:"updated_at"`
	RelatedArticles []DocArticleRelatedResponse `json:"related_articles"`
	TableOfContents []DocTOCItem                `json:"table_of_contents"`
}

// DocArticleRelatedResponse - related article info
type DocArticleRelatedResponse struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Slug            string `json:"slug"`
	Excerpt         string `json:"excerpt"`
	ReadTimeMinutes int    `json:"read_time_minutes"`
	Difficulty      string `json:"difficulty"`
}

// DocTOCItem - table of contents item
type DocTOCItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Level int    `json:"level"`
}

// DocArticleListResponse - paginated articles response
type DocArticleListResponse struct {
	Articles   []DocArticleBriefResponse `json:"articles"`
	Pagination PaginationResponse        `json:"pagination"`
}

// ==================================================
// DOCUMENTATION VIDEO DTOs
// ==================================================

// DocVideoListRequest - request for listing videos
type DocVideoListRequest struct {
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
	Category   string `form:"category"`
	Difficulty string `form:"difficulty"`
}

// DocVideoResponse - video response
type DocVideoResponse struct {
	ID              string                    `json:"id"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description"`
	VideoURL        string                    `json:"video_url"`
	ThumbnailURL    string                    `json:"thumbnail_url"`
	DurationSeconds int                       `json:"duration_seconds"`
	Difficulty      string                    `json:"difficulty"`
	Views           int                       `json:"views"`
	IsFeatured      bool                      `json:"is_featured"`
	Category        *DocCategoryBriefResponse `json:"category,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
}

// DocVideoListResponse - paginated videos response
type DocVideoListResponse struct {
	Videos     []DocVideoResponse `json:"videos"`
	Pagination PaginationResponse `json:"pagination"`
}

// ==================================================
// SEARCH DTOs
// ==================================================

// DocSearchRequest - search request
type DocSearchRequest struct {
	Query string `form:"q" binding:"required,min=2"`
	Type  string `form:"type"` // all, articles, videos
	Limit int    `form:"limit"`
}

// DocSearchResponse - search results
type DocSearchResponse struct {
	Articles     []DocArticleBriefResponse `json:"articles"`
	Videos       []DocVideoResponse        `json:"videos"`
	TotalResults int                       `json:"total_results"`
}

// DocViewResponse - view increment response
type DocViewResponse struct {
	Views int `json:"views"`
}
