package web

import (
	"RyanForce/config"
	"RyanForce/models"
	"RyanForce/utils"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

// ListTechs handles GET /admin/techs
// Displays all technician accounts, optionally filtering by skill.
func ListTechs(c *gin.Context) {
	skill := strings.ToLower(strings.TrimSpace(c.Query("skill")))
	var techs []models.User

	query := config.DB.Where("role = ?", "tech")
	if skill != "" {
		likePattern := "%" + skill + "%"
		query = query.Where("LOWER(skills) LIKE ?", likePattern)
	}

	if err := query.Find(&techs).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error fetching technicians")
		return
	}

	c.HTML(http.StatusOK, "admin_techs_list.html", gin.H{
		"techs": techs,
		"skill": skill,
	})
}

// NewTechForm handles GET /admin/techs/new
// Displays the form to create a new technician.
func NewTechForm(c *gin.Context) {
	c.HTML(http.StatusOK, "admin_tech_new.html", nil)
}

// CreateTech handles POST /admin/techs
// Processes the creation of a new technician account.
func CreateTech(c *gin.Context) {
	var user models.User
	user.Email = c.PostForm("Email")
	user.Name = c.PostForm("Name")
	user.Role = "tech"

	// Parse submitted skills JSON array
	rawSkills := c.PostForm("Skills")
	var skillList []string
	if err := json.Unmarshal([]byte(rawSkills), &skillList); err != nil {
		c.String(http.StatusBadRequest, "Invalid skills format")
		return
	}
	skillsJSON, err := json.Marshal(skillList)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to save skills")
		return
	}
	user.Skills = string(skillsJSON)

	// Validate and hash password
	password := c.PostForm("Password")
	if !utils.IsValidPassword(password) {
		c.String(http.StatusBadRequest, "Password does not meet security requirements")
		return
	}
	hash, err := utils.HashPassword(password)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to hash password")
		return
	}
	user.PasswordHash = hash

	if err := config.DB.Create(&user).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to create technician")
		return
	}

	utils.LogInfo(fmt.Sprintf("[AdminTech] Technician %s created successfully", user.Email))
	c.SetCookie("flash", "Technician created successfully", 3, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/admin/techs")
}

// ShowTech handles GET /admin/techs/:id
// Displays the details for a specific technician.
func ShowTech(c *gin.Context) {
	id := c.Param("id")
	var tech models.User
	if err := config.DB.First(&tech, id).Error; err != nil {
		c.String(http.StatusNotFound, "Technician not found")
		return
	}
	c.HTML(http.StatusOK, "admin_tech_show.html", gin.H{"tech": tech})
}

// EditTechForm handles GET /admin/techs/:id/edit
// Displays the form to edit an existing technician.
func EditTechForm(c *gin.Context) {
	id := c.Param("id")
	var tech models.User
	if err := config.DB.First(&tech, id).Error; err != nil {
		c.String(http.StatusNotFound, "Technician not found")
		return
	}
	c.HTML(http.StatusOK, "admin_tech_edit.html", gin.H{"tech": tech})
}

// UpdateTech handles POST /admin/techs/:id
// Updates an existing technician record.
func UpdateTech(c *gin.Context) {
	id := c.Param("id")
	var tech models.User
	if err := config.DB.First(&tech, id).Error; err != nil {
		c.String(http.StatusNotFound, "Technician not found")
		return
	}

	tech.Email = c.PostForm("Email")
	tech.Name = c.PostForm("Name")

	// Parse updated skills JSON array
	rawSkills := c.PostForm("Skills")
	var skillList []string
	if err := json.Unmarshal([]byte(rawSkills), &skillList); err != nil {
		c.String(http.StatusBadRequest, "Invalid skills format")
		return
	}
	skillsJSON, err := json.Marshal(skillList)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to update skills")
		return
	}
	tech.Skills = string(skillsJSON)

	if err := config.DB.Save(&tech).Error; err != nil {
		c.String(http.StatusInternalServerError, "Failed to update technician")
		return
	}

	utils.LogInfo(fmt.Sprintf("[AdminTech] Technician %s updated successfully", tech.Email))
	c.SetCookie("flash", "Technician updated successfully", 3, "/", "", false, true)
	c.Redirect(http.StatusSeeOther, "/admin/techs")
}

// DeleteTech handles POST /admin/techs/:id/delete
// Deletes a technician record.
func DeleteTech(c *gin.Context) {
	id := c.Param("id")
	var tech models.User
	if err := config.DB.Delete(&tech, id).Error; err != nil {
		c.String(http.StatusInternalServerError, "Error deleting technician")
		return
	}
	c.Redirect(http.StatusFound, "/admin/techs")
}
