package services

import (
	"nexfi-backend/api/repositories"
	"nexfi-backend/dto"
	"nexfi-backend/models"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// DocService handles documentation business logic
type DocService struct {
	repo *repositories.DocRepository
}

// NewDocService creates a new DocService
func NewDocService(db *gorm.DB) *DocService {
	return &DocService{
		repo: repositories.NewDocRepository(db),
	}
}

// ==================================================
// CATEGORY OPERATIONS
// ==================================================

// GetCategories returns all categories with article counts
func (s *DocService) GetCategories() ([]dto.DocCategoryResponse, error) {
	categories, err := s.repo.FindAllCategories()
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DocCategoryResponse, len(categories))
	for i, cat := range categories {
		responses[i] = dto.DocCategoryResponse{
			ID:            cat.ID,
			Name:          cat.Name,
			Slug:          cat.Slug,
			Description:   cat.Description,
			Icon:          cat.Icon,
			Color:         cat.Color,
			Order:         cat.Order,
			ArticlesCount: int(s.repo.GetCategoryArticleCount(cat.ID)),
		}
	}

	return responses, nil
}

// GetCategoryBySlug returns category with its articles
func (s *DocService) GetCategoryBySlug(slug string) (*dto.DocCategoryDetailResponse, error) {
	category, err := s.repo.FindCategoryBySlug(slug)
	if err != nil {
		return nil, err
	}

	articles := make([]dto.DocArticleBriefResponse, len(category.Articles))
	for i, article := range category.Articles {
		articles[i] = s.toArticleBriefResponse(&article)
	}

	return &dto.DocCategoryDetailResponse{
		ID:            category.ID,
		Name:          category.Name,
		Slug:          category.Slug,
		Description:   category.Description,
		Icon:          category.Icon,
		Color:         category.Color,
		Articles:      articles,
		ArticlesCount: len(articles),
	}, nil
}

// ==================================================
// ARTICLE OPERATIONS
// ==================================================

// ListArticles returns articles with filters and pagination
func (s *DocService) ListArticles(req dto.DocArticleListRequest) (*dto.DocArticleListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	filter := repositories.ArticleFilter{
		CategorySlug: req.Category,
		Difficulty:   req.Difficulty,
		Search:       req.Search,
		Sort:         req.Sort,
	}

	articles, total, err := s.repo.FindAllArticles(filter, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DocArticleBriefResponse, len(articles))
	for i, article := range articles {
		responses[i] = s.toArticleBriefResponse(&article)
	}

	return &dto.DocArticleListResponse{
		Articles: responses,
		Pagination: dto.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      int(total),
			TotalPages: int((total + int64(req.Limit) - 1) / int64(req.Limit)),
		},
	}, nil
}

// GetArticleBySlug returns full article details
func (s *DocService) GetArticleBySlug(slug string) (*dto.DocArticleDetailResponse, error) {
	article, err := s.repo.FindArticleBySlug(slug)
	if err != nil {
		return nil, err
	}

	// Get related articles
	var tags []string
	if article.Tags != nil {
		tags = article.Tags
	}
	relatedArticles, _ := s.repo.GetRelatedArticles(article.ID, article.CategoryID, tags, 5)
	related := make([]dto.DocArticleRelatedResponse, len(relatedArticles))
	for i, ra := range relatedArticles {
		related[i] = dto.DocArticleRelatedResponse{
			ID:              ra.ID,
			Title:           ra.Title,
			Slug:            ra.Slug,
			Excerpt:         ra.Excerpt,
			ReadTimeMinutes: ra.ReadTimeMinutes,
			Difficulty:      string(ra.Difficulty),
		}
	}

	// Generate table of contents from markdown
	toc := s.generateTableOfContents(article.Content)

	response := &dto.DocArticleDetailResponse{
		ID:              article.ID,
		Title:           article.Title,
		Slug:            article.Slug,
		Excerpt:         article.Excerpt,
		Content:         article.Content,
		ReadTimeMinutes: article.ReadTimeMinutes,
		Difficulty:      string(article.Difficulty),
		Tags:            article.Tags,
		Views:           article.Views,
		IsFeatured:      article.IsFeatured,
		PublishedAt:     article.PublishedAt,
		UpdatedAt:       article.UpdatedAt,
		RelatedArticles: related,
		TableOfContents: toc,
	}

	if article.Category != nil {
		response.Category = &dto.DocCategoryBriefResponse{
			ID:   article.Category.ID,
			Name: article.Category.Name,
			Slug: article.Category.Slug,
		}
	}

	if article.Author != nil {
		response.Author = &dto.DocAuthorResponse{
			ID:        article.Author.ID,
			Name:      article.Author.Name,
			AvatarURL: article.Author.AvatarURL,
		}
	}

	return response, nil
}

// GetPopularArticles returns most viewed articles
func (s *DocService) GetPopularArticles(limit int) ([]dto.DocArticleBriefResponse, error) {
	if limit < 1 {
		limit = 10
	}

	articles, err := s.repo.GetPopularArticles(limit)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DocArticleBriefResponse, len(articles))
	for i, article := range articles {
		responses[i] = s.toArticleBriefResponse(&article)
	}

	return responses, nil
}

// GetFeaturedArticles returns featured articles
func (s *DocService) GetFeaturedArticles(limit int) ([]dto.DocArticleBriefResponse, error) {
	if limit < 1 {
		limit = 10
	}

	articles, err := s.repo.GetFeaturedArticles(limit)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DocArticleBriefResponse, len(articles))
	for i, article := range articles {
		responses[i] = s.toArticleBriefResponse(&article)
	}

	return responses, nil
}

// IncrementArticleView increments article view count
func (s *DocService) IncrementArticleView(slug string) (*dto.DocViewResponse, error) {
	views, err := s.repo.IncrementArticleViews(slug)
	if err != nil {
		return nil, err
	}
	return &dto.DocViewResponse{Views: views}, nil
}

// ==================================================
// VIDEO OPERATIONS
// ==================================================

// ListVideos returns videos with filters and pagination
func (s *DocService) ListVideos(req dto.DocVideoListRequest) (*dto.DocVideoListResponse, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	filter := repositories.VideoFilter{
		CategorySlug: req.Category,
		Difficulty:   req.Difficulty,
	}

	videos, total, err := s.repo.FindAllVideos(filter, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DocVideoResponse, len(videos))
	for i, video := range videos {
		responses[i] = s.toVideoResponse(&video)
	}

	return &dto.DocVideoListResponse{
		Videos: responses,
		Pagination: dto.PaginationResponse{
			Page:       req.Page,
			Limit:      req.Limit,
			Total:      int(total),
			TotalPages: int((total + int64(req.Limit) - 1) / int64(req.Limit)),
		},
	}, nil
}

// GetVideoByID returns video details
func (s *DocService) GetVideoByID(id string) (*dto.DocVideoResponse, error) {
	video, err := s.repo.FindVideoByID(id)
	if err != nil {
		return nil, err
	}

	// Increment views
	s.repo.IncrementVideoViews(id)

	response := s.toVideoResponse(video)
	return &response, nil
}

// ==================================================
// SEARCH OPERATIONS
// ==================================================

// Search searches articles and videos
func (s *DocService) Search(req dto.DocSearchRequest) (*dto.DocSearchResponse, error) {
	if req.Limit < 1 {
		req.Limit = 10
	}
	if req.Type == "" {
		req.Type = "all"
	}

	response := &dto.DocSearchResponse{
		Articles: []dto.DocArticleBriefResponse{},
		Videos:   []dto.DocVideoResponse{},
	}

	// Search articles
	if req.Type == "all" || req.Type == "articles" {
		articles, err := s.repo.SearchArticles(req.Query, req.Limit)
		if err == nil {
			for _, article := range articles {
				response.Articles = append(response.Articles, s.toArticleBriefResponse(&article))
			}
		}
	}

	// Search videos
	if req.Type == "all" || req.Type == "videos" {
		videos, err := s.repo.SearchVideos(req.Query, req.Limit)
		if err == nil {
			for _, video := range videos {
				response.Videos = append(response.Videos, s.toVideoResponse(&video))
			}
		}
	}

	response.TotalResults = len(response.Articles) + len(response.Videos)

	return response, nil
}

// ==================================================
// HELPER FUNCTIONS
// ==================================================

func (s *DocService) toArticleBriefResponse(article *models.DocArticle) dto.DocArticleBriefResponse {
	response := dto.DocArticleBriefResponse{
		ID:              article.ID,
		Title:           article.Title,
		Slug:            article.Slug,
		Excerpt:         article.Excerpt,
		ReadTimeMinutes: article.ReadTimeMinutes,
		Difficulty:      string(article.Difficulty),
		Tags:            article.Tags,
		Views:           article.Views,
		IsFeatured:      article.IsFeatured,
		PublishedAt:     article.PublishedAt,
	}

	// Check if article is new (published within last 7 days)
	if article.PublishedAt != nil {
		response.IsNew = time.Since(*article.PublishedAt).Hours() < 168 // 7 days
	}

	if article.Category != nil {
		response.Category = &dto.DocCategoryBriefResponse{
			ID:   article.Category.ID,
			Name: article.Category.Name,
			Slug: article.Category.Slug,
		}
	}

	if article.Author != nil {
		response.Author = &dto.DocAuthorResponse{
			ID:        article.Author.ID,
			Name:      article.Author.Name,
			AvatarURL: article.Author.AvatarURL,
		}
	}

	return response
}

func (s *DocService) toVideoResponse(video *models.DocVideo) dto.DocVideoResponse {
	response := dto.DocVideoResponse{
		ID:              video.ID,
		Title:           video.Title,
		Description:     video.Description,
		VideoURL:        video.VideoURL,
		ThumbnailURL:    video.ThumbnailURL,
		DurationSeconds: video.DurationSeconds,
		Difficulty:      string(video.Difficulty),
		Views:           video.Views,
		IsFeatured:      video.IsFeatured,
		CreatedAt:       video.CreatedAt,
	}

	if video.Category != nil {
		response.Category = &dto.DocCategoryBriefResponse{
			Name: video.Category.Name,
			Slug: video.Category.Slug,
		}
	}

	return response
}

// generateTableOfContents extracts headings from markdown content
func (s *DocService) generateTableOfContents(content string) []dto.DocTOCItem {
	var toc []dto.DocTOCItem

	// Match markdown headings (# Heading, ## Heading, ### Heading)
	re := regexp.MustCompile(`(?m)^(#{1,3})\s+(.+)$`)
	matches := re.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		level := len(match[1])
		title := strings.TrimSpace(match[2])
		id := strings.ToLower(strings.ReplaceAll(title, " ", "-"))
		// Remove special characters from ID
		id = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(id, "")

		toc = append(toc, dto.DocTOCItem{
			ID:    id,
			Title: title,
			Level: level,
		})
	}

	return toc
}
