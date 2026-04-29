package model

import (
	"errors"
	"strings"
)

const (
	SkillStatusEnabled  = 1
	SkillStatusDisabled = 2
)

type Skill struct {
	Id          int    `json:"id" gorm:"column:id;type:int;primaryKey;autoIncrement"`
	Name        string `json:"name" gorm:"column:name;type:varchar(100);not null"`
	Description string `json:"description" gorm:"column:description;type:text"`
	Detail      string `json:"detail" gorm:"column:detail;type:text"`
	Category    string `json:"category" gorm:"column:category;type:varchar(20);not null;index"`
	ImageUrl    string `json:"image_url" gorm:"column:image_url;type:varchar(255)"`
	FileUrl     string `json:"file_url" gorm:"column:file_url;type:varchar(255)"`
	FileName    string `json:"file_name" gorm:"column:file_name;type:varchar(100)"`
	Version     string `json:"version" gorm:"column:version;type:varchar(20);default:'1.0.0'"`
	Author      string `json:"author" gorm:"column:author;type:varchar(50)"`
	Status      int    `json:"status" gorm:"column:status;type:int;default:1;index"`
	Downloads   int    `json:"downloads" gorm:"column:downloads;type:int;default:0"`
	CreatedAt   int64  `json:"created_at" gorm:"column:created_at;type:bigint;autoCreateTime"`
	UpdatedAt   int64  `json:"updated_at" gorm:"column:updated_at;type:bigint;autoUpdateTime"`
}

func (Skill) TableName() string {
	return "skills"
}

func GetAllSkills(startIdx int, num int) ([]*Skill, error) {
	var skills []*Skill
	err := DB.Where("status = ?", SkillStatusEnabled).Order("id DESC").Limit(num).Offset(startIdx).Find(&skills).Error
	return skills, err
}

func GetSkillById(id int) (*Skill, error) {
	var skill Skill
	err := DB.Where("id = ? AND status = ?", id, SkillStatusEnabled).First(&skill).Error
	return &skill, err
}

func SearchSkills(keyword string, category string) ([]*Skill, error) {
	var skills []*Skill
	query := DB.Where("status = ?", SkillStatusEnabled)
	if keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if category != "" && category != "all" {
		query = query.Where("category = ?", category)
	}
	err := query.Order("id DESC").Find(&skills).Error
	return skills, err
}

func GetAllSkillsForAdmin(startIdx int, num int) ([]*Skill, error) {
	var skills []*Skill
	err := DB.Order("id DESC").Limit(num).Offset(startIdx).Find(&skills).Error
	return skills, err
}

func GetSkillByIdForAdmin(id int) (*Skill, error) {
	var skill Skill
	err := DB.Where("id = ?", id).First(&skill).Error
	return &skill, err
}

func InsertSkill(skill *Skill) error {
	var err error
	err = DB.Create(skill).Error
	return err
}

func UpdateSkill(skill *Skill) error {
	var err error
	err = DB.Model(skill).Select("name", "description", "detail", "category", "image_url", "file_url", "file_name", "version", "author", "status").Updates(skill).Error
	return err
}

func DeleteSkill(id int) error {
	if id == 0 {
		return errors.New("id 为空！")
	}
	err := DB.Where("id = ?", id).Delete(&Skill{}).Error
	return err
}

func GetSkillsByCategory(category string) ([]*Skill, error) {
	var skills []*Skill
	err := DB.Where("status = ? AND category = ?", SkillStatusEnabled, category).Order("id DESC").Find(&skills).Error
	return skills, err
}

func CountSkills() (int64, error) {
	var count int64
	err := DB.Model(&Skill{}).Where("status = ?", SkillStatusEnabled).Count(&count).Error
	return count, err
}

func CountAllSkills() (int64, error) {
	var count int64
	err := DB.Model(&Skill{}).Count(&count).Error
	return count, err
}

// SkillCategory represents a user-managed skill category
type SkillCategory struct {
	Id   int    `json:"id" gorm:"column:id;type:int;primaryKey;autoIncrement"`
	Name string `json:"name" gorm:"column:name;type:varchar(64);not null;uniqueIndex"`
	Sort int    `json:"sort" gorm:"column:sort;type:int;default:0"`
}

func (SkillCategory) TableName() string {
	return "skill_categories"
}

// GetAllSkillCategories returns all skill categories ordered by sort
func GetAllSkillCategories() ([]*SkillCategory, error) {
	var categories []*SkillCategory
	err := DB.Order("sort asc, id asc").Find(&categories).Error
	return categories, err
}

// CreateSkillCategory creates a new skill category
func CreateSkillCategory(category *SkillCategory) error {
	category.Name = strings.TrimSpace(category.Name)
	if category.Name == "" {
		return errors.New("分类名称不能为空")
	}
	// Check duplicate
	var count int64
	DB.Model(&SkillCategory{}).Where("name = ?", category.Name).Count(&count)
	if count > 0 {
		return errors.New("分类名称已存在")
	}
	return DB.Create(category).Error
}

// UpdateSkillCategory updates a skill category
func UpdateSkillCategory(category *SkillCategory) error {
	category.Name = strings.TrimSpace(category.Name)
	if category.Name == "" {
		return errors.New("分类名称不能为空")
	}
	// Check duplicate (exclude self)
	var count int64
	DB.Model(&SkillCategory{}).Where("name = ? AND id != ?", category.Name, category.Id).Count(&count)
	if count > 0 {
		return errors.New("分类名称已存在")
	}
	return DB.Model(&SkillCategory{}).Where("id = ?", category.Id).Updates(category).Error
}

// DeleteSkillCategory deletes a skill category
func DeleteSkillCategory(id int) error {
	return DB.Delete(&SkillCategory{}, id).Error
}

// InitDefaultSkillCategories initializes default categories if table is empty
func InitDefaultSkillCategories() error {
	var count int64
	if err := DB.Model(&SkillCategory{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	defaults := []string{"产品经理类", "开发类", "测试类", "通用类"}
	for i, name := range defaults {
		cat := &SkillCategory{
			Name: name,
			Sort: i,
		}
		if err := DB.Create(cat).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetSkillCategoryNamesMap returns all category names as a map key->displayName
func GetSkillCategoryNamesMap() (map[string]string, error) {
	categories, err := GetAllSkillCategories()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string, len(categories))
	for _, c := range categories {
		result[c.Name] = c.Name
	}
	return result, nil
}

// UpdateSkillFileInfo updates the file info (url + name) of a skill
func UpdateSkillFileInfo(id int, fileUrl string, fileName string) error {
	return DB.Model(&Skill{}).Where("id = ?", id).Updates(map[string]interface{}{
		"file_url":  fileUrl,
		"file_name": fileName,
	}).Error
}

// IncrementSkillDownloads increments the download count of a skill
func IncrementSkillDownloads(id int) error {
	var skill Skill
	err := DB.First(&skill, id).Error
	if err != nil {
		return err
	}
	return DB.Model(&Skill{}).Where("id = ?", id).Update("downloads", skill.Downloads+1).Error
}
