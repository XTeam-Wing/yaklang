package schema

import (
	"github.com/jinzhu/gorm"
	"github.com/yaklang/yaklang/common/utils"
	"github.com/yaklang/yaklang/common/utils/bizhelper"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
)

type AIYakTool struct {
	gorm.Model

	Name        string `json:"name" gorm:"unique_index"`
	Description string `json:"description" gorm:"type:text;index"`
	Keywords    string `json:"keywords" gorm:"type:text;index"`
	Content     string `json:"content" gorm:"type:text"`
	Params      string `json:"params" gorm:"type:text"`
	Path        string `json:"path" gorm:"type:text;index"`
	Hash        string `json:"hash"`
	IsFavorite  bool   `json:"is_favorite" gorm:"default:false;index"`
}

func (*AIYakTool) TableName() string {
	return "ai_yak_tools"
}

func (d *AIYakTool) CalcHash() string {
	return utils.CalcSha1(d.Name, d.Content, d.Params, d.Path, d.Description, d.Keywords)
}

func (d *AIYakTool) BeforeSave() error {
	d.Hash = d.CalcHash()
	return nil
}

func SaveAIYakTool(db *gorm.DB, tool *AIYakTool) (int64, error) {
	db = db.Model(&AIYakTool{})
	if db := db.Where("name = ?", tool.Name).Assign(tool).FirstOrCreate(&AIYakTool{}); db.Error != nil {
		return 0, utils.Errorf("create/update AIYakTool failed: %s", db.Error)
	}
	return db.RowsAffected, nil
}

func GetAIYakTool(db *gorm.DB, name string) (*AIYakTool, error) {
	db = db.Model(&AIYakTool{})
	var tool AIYakTool
	if err := db.Where("name = ?", name).First(&tool).Error; err != nil {
		return nil, err
	}
	return &tool, nil
}
func SearchAIYakToolByPath(db *gorm.DB, path string) ([]*AIYakTool, error) {
	db = db.Model(&AIYakTool{})
	var tools []*AIYakTool
	db = bizhelper.FuzzSearchEx(db, []string{"path"}, path, false)
	if err := db.Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}
func SearchAIYakTool(db *gorm.DB, keywords string) ([]*AIYakTool, error) {
	db = db.Model(&AIYakTool{})
	var tools []*AIYakTool
	if keywords != "" {
		db = bizhelper.FuzzSearchEx(db, []string{"name", "keywords", "description", "path"}, keywords, false)
	}

	if err := db.Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

func DeleteAIYakTools(db *gorm.DB, names ...string) (int64, error) {
	db = db.Model(&AIYakTool{})
	if db := db.Where("name IN (?)", names).Delete(&AIYakTool{}); db.Error != nil {
		return 0, utils.Errorf("delete AIYakTool failed: %s", db.Error)
	}
	return db.RowsAffected, nil
}

// SearchAIYakToolWithPagination adds pagination support to AIYakTool search
func SearchAIYakToolWithPagination(db *gorm.DB, keywords string, onlyFavorites bool, paging *ypb.Paging) (*bizhelper.Paginator, []*AIYakTool, error) {
	orderBy := paging.GetOrderBy()
	order := paging.GetOrder()
	page := int(paging.GetPage())
	limit := int(paging.GetLimit())

	db = db.Model(&AIYakTool{})

	// Apply fuzzy search if keywords provided
	if keywords != "" {
		db = bizhelper.FuzzSearchEx(db, []string{"name", "keywords", "description", "path"}, keywords, false)
	}

	// Apply favorite filter if requested
	if onlyFavorites {
		db = db.Where("is_favorite = ?", true)
	}

	// Apply ordering
	if orderBy == "" {
		orderBy = "updated_at"
	}
	if order == "" {
		order = "desc"
	}
	db = bizhelper.QueryOrder(db, orderBy, order)

	// Perform paginated query
	var tools []*AIYakTool
	paginator, db := bizhelper.Paging(db, page, limit, &tools)
	if db.Error != nil {
		return nil, nil, utils.Errorf("search AIYakTool with pagination failed: %s", db.Error)
	}

	return paginator, tools, nil
}

// ToggleAIYakToolFavorite toggles the favorite status of an AI tool
func ToggleAIYakToolFavorite(db *gorm.DB, toolName string) (bool, error) {
	db = db.Model(&AIYakTool{})

	var tool AIYakTool
	if err := db.Where("name = ?", toolName).First(&tool).Error; err != nil {
		return false, utils.Errorf("AI tool not found: %s", err)
	}

	// Toggle the favorite status
	tool.IsFavorite = !tool.IsFavorite

	if err := db.Save(&tool).Error; err != nil {
		return false, utils.Errorf("failed to update AI tool favorite status: %s", err)
	}

	return tool.IsFavorite, nil
}
