package repositories

import (
	"nexfi-backend/models"

	"gorm.io/gorm"
)

// DocRepository handles documentation database operations
type DocRepository struct {
	*BaseRepository
}

// NewDocRepository creates a new DocRepository
func NewDocRepository(db *gorm.DB) *DocRepository {
	return &DocRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// ==================================================
// CATEGORY OPERATIONS
// ==================================================

// FindAllCategories finds all active categories with article counts
func (r *DocRepository) FindAllCategories() ([]models.DocCategory, error) {
	var categories []models.DocCategory
	err := r.DB.Where("is_active = ?", true).
		Order("\"order\" ASC, name ASC").
		Find(&categories).Error
	return categories, err
}

// FindCategoryBySlug finds category by slug with articles
func (r *DocRepository) FindCategoryBySlug(slug string) (*models.DocCategory, error) {
	var category models.DocCategory
	err := r.DB.Where("slug = ? AND is_active = ?", slug, true).
		Preload("Articles", func(db *gorm.DB) *gorm.DB {
			return db.Where("is_published = ?", true).
				Preload("Author").
				Order("\"order\" ASC, published_at DESC")
		}).
		First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

// GetCategoryArticleCount returns count of published articles in a category
func (r *DocRepository) GetCategoryArticleCount(categoryID string) int64 {
	var count int64
	r.DB.Model(&models.DocArticle{}).
		Where("category_id = ? AND is_published = ?", categoryID, true).
		Count(&count)
	return count
}

// ==================================================
// ARTICLE OPERATIONS
// ==================================================

// ArticleFilter defines filter options for articles
type ArticleFilter struct {
	CategorySlug string
	Difficulty   string
	Search       string
	Sort         string // newest, popular, title
	IsFeatured   bool
}

// FindAllArticles finds articles with filters and pagination
func (r *DocRepository) FindAllArticles(filter ArticleFilter, page, limit int) ([]models.DocArticle, int64, error) {
	var articles []models.DocArticle
	var total int64

	query := r.DB.Model(&models.DocArticle{}).Where("is_published = ?", true)

	// Category filter
	if filter.CategorySlug != "" {
		query = query.Joins("JOIN doc_categories ON doc_categories.id = doc_articles.category_id").
			Where("doc_categories.slug = ?", filter.CategorySlug)
	}

	// Difficulty filter
	if filter.Difficulty != "" {
		query = query.Where("doc_articles.difficulty = ?", filter.Difficulty)
	}

	// Search filter
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("doc_articles.title ILIKE ? OR doc_articles.excerpt ILIKE ? OR doc_articles.content ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	// Featured filter
	if filter.IsFeatured {
		query = query.Where("doc_articles.is_featured = ?", true)
	}

	query.Count(&total)

	// Sorting
	switch filter.Sort {
	case "popular":
		query = query.Order("doc_articles.views DESC")
	case "title":
		query = query.Order("doc_articles.title ASC")
	default: // newest
		query = query.Order("doc_articles.published_at DESC")
	}

	err := query.Scopes(Paginate(page, limit)).
		Preload("Category").
		Preload("Author").
		Find(&articles).Error

	return articles, total, err
}

// FindArticleBySlug finds article by slug
func (r *DocRepository) FindArticleBySlug(slug string) (*models.DocArticle, error) {
	var article models.DocArticle
	err := r.DB.Where("slug = ? AND is_published = ?", slug, true).
		Preload("Category").
		Preload("Author").
		First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// GetPopularArticles gets most viewed articles
func (r *DocRepository) GetPopularArticles(limit int) ([]models.DocArticle, error) {
	var articles []models.DocArticle
	err := r.DB.Where("is_published = ?", true).
		Preload("Category").
		Order("views DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

// GetFeaturedArticles gets featured articles
func (r *DocRepository) GetFeaturedArticles(limit int) ([]models.DocArticle, error) {
	var articles []models.DocArticle
	err := r.DB.Where("is_published = ? AND is_featured = ?", true, true).
		Preload("Category").
		Preload("Author").
		Order("published_at DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

// GetRelatedArticles finds related articles based on category and tags
func (r *DocRepository) GetRelatedArticles(articleID, categoryID string, tags []string, limit int) ([]models.DocArticle, error) {
	var articles []models.DocArticle

	query := r.DB.Where("id != ? AND is_published = ?", articleID, true)

	if len(tags) > 0 {
		// Articles with matching tags get priority (simplified - using category match)
		query = query.Where("category_id = ? OR tags && ?", categoryID, tags)
	} else {
		query = query.Where("category_id = ?", categoryID)
	}

	err := query.Order("views DESC").
		Limit(limit).
		Find(&articles).Error

	return articles, err
}

// IncrementArticleViews increments view count
func (r *DocRepository) IncrementArticleViews(slug string) (int, error) {
	var article models.DocArticle
	err := r.DB.Model(&models.DocArticle{}).
		Where("slug = ?", slug).
		UpdateColumn("views", gorm.Expr("views + 1")).
		Error
	if err != nil {
		return 0, err
	}

	r.DB.Where("slug = ?", slug).Select("views").First(&article)
	return article.Views, nil
}

// ==================================================
// VIDEO OPERATIONS
// ==================================================

// VideoFilter defines filter options for videos
type VideoFilter struct {
	CategorySlug string
	Difficulty   string
}

// FindAllVideos finds videos with filters and pagination
func (r *DocRepository) FindAllVideos(filter VideoFilter, page, limit int) ([]models.DocVideo, int64, error) {
	var videos []models.DocVideo
	var total int64

	query := r.DB.Model(&models.DocVideo{}).Where("is_published = ?", true)

	if filter.CategorySlug != "" {
		query = query.Joins("JOIN doc_categories ON doc_categories.id = doc_videos.category_id").
			Where("doc_categories.slug = ?", filter.CategorySlug)
	}

	if filter.Difficulty != "" {
		query = query.Where("doc_videos.difficulty = ?", filter.Difficulty)
	}

	query.Count(&total)

	err := query.Scopes(Paginate(page, limit)).
		Preload("Category").
		Order("doc_videos.created_at DESC").
		Find(&videos).Error

	return videos, total, err
}

// FindVideoByID finds video by ID
func (r *DocRepository) FindVideoByID(id string) (*models.DocVideo, error) {
	var video models.DocVideo
	err := r.DB.Where("id = ? AND is_published = ?", id, true).
		Preload("Category").
		First(&video).Error
	if err != nil {
		return nil, err
	}
	return &video, nil
}

// IncrementVideoViews increments video view count
func (r *DocRepository) IncrementVideoViews(id string) error {
	return r.DB.Model(&models.DocVideo{}).
		Where("id = ?", id).
		UpdateColumn("views", gorm.Expr("views + 1")).
		Error
}

// ==================================================
// SEARCH OPERATIONS
// ==================================================

// SearchArticles searches articles by query
func (r *DocRepository) SearchArticles(query string, limit int) ([]models.DocArticle, error) {
	var articles []models.DocArticle
	searchPattern := "%" + query + "%"

	err := r.DB.Where("is_published = ?", true).
		Where("title ILIKE ? OR excerpt ILIKE ? OR content ILIKE ?", searchPattern, searchPattern, searchPattern).
		Preload("Category").
		Order("views DESC").
		Limit(limit).
		Find(&articles).Error

	return articles, err
}

// SearchVideos searches videos by query
func (r *DocRepository) SearchVideos(query string, limit int) ([]models.DocVideo, error) {
	var videos []models.DocVideo
	searchPattern := "%" + query + "%"

	err := r.DB.Where("is_published = ?", true).
		Where("title ILIKE ? OR description ILIKE ?", searchPattern, searchPattern).
		Preload("Category").
		Order("views DESC").
		Limit(limit).
		Find(&videos).Error

	return videos, err
}
