package controller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// GetSkills returns all skills for marketplace (public)
func GetSkills(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	startIdx := (page - 1) * pageSize

	skills, err := model.GetAllSkills(startIdx, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取技能列表失败",
		})
		return
	}

	total, _ := model.CountSkills()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    skills,
		"total":   total,
	})
}

// GetSkill returns a single skill by id (public)
func GetSkill(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的技能ID",
		})
		return
	}

	skill, err := model.GetSkillById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "技能不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    skill,
	})
}

// SearchSkills searches skills by keyword and category (public)
func SearchSkills(c *gin.Context) {
	keyword := c.Query("keyword")
	category := c.Query("category")

	skills, err := model.SearchSkills(keyword, category)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "搜索失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    skills,
	})
}

// GetSkillCategories returns skill category names map (public)
func GetSkillCategories(c *gin.Context) {
	names, err := model.GetSkillCategoryNamesMap()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取分类失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    names,
	})
}

// AdminGetSkills returns all skills for admin
func AdminGetSkills(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	startIdx := (page - 1) * pageSize

	skills, err := model.GetAllSkillsForAdmin(startIdx, pageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取技能列表失败",
		})
		return
	}

	total, _ := model.CountAllSkills()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    skills,
		"total":   total,
	})
}

// AdminGetSkill returns a single skill for admin
func AdminGetSkill(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的技能ID",
		})
		return
	}

	skill, err := model.GetSkillByIdForAdmin(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "技能不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    skill,
	})
}

// AdminCreateSkill creates a new skill
func AdminCreateSkill(c *gin.Context) {
	var skill model.Skill
	if err := common.DecodeJson(c.Request.Body, &skill); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	if skill.Name == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "技能名称不能为空",
		})
		return
	}
	if skill.Category == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "分类不能为空",
		})
		return
	}

	err := model.InsertSkill(&skill)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "创建技能失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    skill,
	})
}

// AdminUpdateSkill updates a skill
func AdminUpdateSkill(c *gin.Context) {
	var skill model.Skill
	if err := common.DecodeJson(c.Request.Body, &skill); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	if skill.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "技能ID不能为空",
		})
		return
	}

	err := model.UpdateSkill(&skill)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "更新技能失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    skill,
	})
}

// AdminDeleteSkill deletes a skill
func AdminDeleteSkill(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的技能ID",
		})
		return
	}

	err = model.DeleteSkill(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "删除技能失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// DownloadSkillFile downloads a skill file (public, increments download count)
func DownloadSkillFile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的技能ID",
		})
		return
	}

	skill, err := model.GetSkillById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "技能不存在",
		})
		return
	}

	if skill.FileUrl == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "该技能没有文件",
		})
		return
	}

	// Increment download count
	_ = model.IncrementSkillDownloads(id)

	// Remove leading slash from FileUrl if present
	fileUrl := skill.FileUrl
	if len(fileUrl) > 0 && fileUrl[0] == '/' {
		fileUrl = fileUrl[1:]
	}
	filePath := filepath.Join(".", fileUrl)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "文件不存在",
		})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", skill.FileName))
	c.File(filePath)
}

// UploadSkillFile uploads a file for a skill (admin, binds to skill_id)
func UploadSkillFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "文件上传失败",
		})
		return
	}

	// Validate file type
	ext := filepath.Ext(file.Filename)
	isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" || ext == ".webp"

	// Validate file size
	maxSize := 10 * 1024 * 1024 // 10MB for skill files
	if isImage {
		maxSize = 2 * 1024 * 1024 // 2MB for images
	}
	if file.Size >= int64(maxSize) {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("文件大小超过 %dMB 限制", maxSize/(1024*1024)),
		})
		return
	}

	// Optional: bind to a specific skill
	skillIdStr := c.PostForm("skill_id")
	var skillId int
	if skillIdStr != "" {
		skillId, _ = strconv.Atoi(skillIdStr)
	}

	// Create upload directory if not exists
	uploadDir := "./uploads/skills"
	if isImage {
		uploadDir = "./uploads/skills/images"
	}
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err = os.MkdirAll(uploadDir, 0755)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "创建上传目录失败",
			})
			return
		}
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s%s", common.GetTimestamp(), common.GetRandomString(8), ext)
	filepathStr := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filepathStr); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "保存文件失败",
		})
		return
	}

	fileUrl := "/uploads/skills/" + filename
	if isImage {
		fileUrl = "/uploads/skills/images/" + filename
	}

	// If skill_id is provided, update the skill's file or image info
	if skillId > 0 {
		skill, err := model.GetSkillByIdForAdmin(skillId)
		if err == nil && skill != nil {
			if isImage {
				if skill.ImageUrl != "" {
					oldPath := filepath.Join(".", skill.ImageUrl)
					os.Remove(oldPath)
				}
				model.DB.Model(&model.Skill{}).Where("id = ?", skillId).Update("image_url", fileUrl)
			} else {
				if skill.FileUrl != "" {
					oldPath := filepath.Join(".", skill.FileUrl)
					os.Remove(oldPath)
				}
				model.UpdateSkillFileInfo(skillId, fileUrl, file.Filename)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"url":      fileUrl,
			"filename": file.Filename,
			"skill_id": skillId,
			"is_image": isImage,
		},
	})
}

// UploadSkillImage uploads an image for a skill icon (admin)
func UploadSkillImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "图片上传失败",
		})
		return
	}

	// Validate image type
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" && ext != ".webp" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "不支持的图片格式，仅支持 jpg, jpeg, png, gif, webp",
		})
		return
	}

	// Create upload directory if not exists
	uploadDir := "./uploads/skills/images"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		err = os.MkdirAll(uploadDir, 0755)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "创建上传目录失败",
			})
			return
		}
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s%s", common.GetTimestamp(), common.GetRandomString(8), ext)
	filepathStr := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filepathStr); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "保存图片失败",
		})
		return
	}

	fileUrl := "/uploads/skills/images/" + filename

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"url":      fileUrl,
			"filename": file.Filename,
		},
	})
}

// AdminGetSkillCategories returns all skill categories for admin
func AdminGetSkillCategories(c *gin.Context) {
	categories, err := model.GetAllSkillCategories()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取分类列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    categories,
	})
}

// AdminCreateSkillCategory creates a new skill category
func AdminCreateSkillCategory(c *gin.Context) {
	var category model.SkillCategory
	if err := common.DecodeJson(c.Request.Body, &category); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	if category.Name == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "分类名称不能为空",
		})
		return
	}

	err := model.CreateSkillCategory(&category)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    category,
	})
}

// AdminUpdateSkillCategory updates a skill category
func AdminUpdateSkillCategory(c *gin.Context) {
	var category model.SkillCategory
	if err := common.DecodeJson(c.Request.Body, &category); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}

	if category.Id == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "分类ID不能为空",
		})
		return
	}

	err := model.UpdateSkillCategory(&category)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    category,
	})
}

// AdminDeleteSkillCategory deletes a skill category
func AdminDeleteSkillCategory(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的分类ID",
		})
		return
	}

	err = model.DeleteSkillCategory(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "删除分类失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
