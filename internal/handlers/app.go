package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type AppHandler struct {
	handler *Handler
}

func NewAppHandler(h *Handler) *AppHandler {
    return &AppHandler{
        handler: h,
    }
}

// HandleLanding handles the landing page
func (h *AppHandler) HandleLanding(c *gin.Context) {
    user := h.getCurrentUser(c)
    fmt.Printf("DEBUG: Landing page - current user: '%s'\n", user)
    
    // Get session info for debugging
    sessionID, _ := c.Cookie("session")
    
    templateData := gin.H{
        "user": user,
        "storage_backend": h.handler.Config.StorageBackend,
        "environment": h.handler.Config.Environment,
        "session_id": sessionID,
        "debug": h.handler.Config.Environment == "development",
    }
    
    c.HTML(http.StatusOK, "landing-page.html", templateData)
}
func (h *AppHandler) HandleAmazonWebApp(c *gin.Context) {
	paramCode := c.Param("paramCode")
	param2 := c.Param("param2")

	// Get current user from cookie
	user := h.getCurrentUser(c)
	if user == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		return
	}

	// For TouchCalc integration
	param1 := ""
	if param1 == "touchcalc" {
		if user != "" {
			if(user == "") {
				c.Redirect(http.StatusTemporaryRedirect, "/login")
				return
			}
			h.handleTouchCalc(c, user, paramCode, param2)
		} else {
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"Error": "Invalid user type",
			})
		}
		return
	}

	// Default handling for other apps
	if user != "" {
		h.handleGenericApp(c, user, param1, paramCode, param2)
	} else {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Invalid user type",
		})
	}
}

func (h *AppHandler) HandleGoogleVerification(c *gin.Context) {
    slug := c.Param("filepath")
    // Remove leading slash
    if len(slug) > 0 && slug[0] == '/' {
        slug = slug[1:]
    }
    
    c.HTML(http.StatusOK, slug, gin.H{})
}

func (h *AppHandler) getCurrentUser(c *gin.Context) string {
    userCookie, err := c.Cookie("user")
    if err != nil {
        return ""
    }

    // Handle both JSON format and plain text format
    if len(userCookie) > 0 && userCookie[0] == '"' && userCookie[len(userCookie)-1] == '"' {
        // JSON format
        var user string
        err = json.Unmarshal([]byte(userCookie), &user)
        if err != nil {
            return ""
        }
        return user
    }
    
    // Plain text format
    return userCookie
}

func (h *AppHandler) handleTouchCalc(c *gin.Context, user, code, filename string) {
	// Load TouchCalc configuration
	configPath := filepath.Join("webappTemplates", "touchcalc", "touchcalc.config.txt")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Printf("Error reading TouchCalc config: %v\n", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Failed to load TouchCalc configuration",
		})
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal(configData, &config); err != nil {
		fmt.Printf("Error parsing TouchCalc config: %v\n", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"Error": "Invalid TouchCalc configuration",
		})
		return
	}

	// Get user's spreadsheet list
	sheets, err := h.getUserSpreadsheets(user, "touchcalc")
	if err != nil {
		fmt.Printf("Error getting user spreadsheets: %v\n", err)
		sheets = []string{"default"} // fallback
	}

	// Determine filename
	fname := filename
	if fname == "" {
		fname = "default"
	}

	// Check Dropbox connection status (if implemented)
	dbLogin := 0 // Default to not connected

	data := gin.H{
		"fname":   fname,
		"user":    user,
		"storage": h.handler.Config.StorageBackend,
		"sheets":  sheets,
		"dbLogin": dbLogin,
		"code":    code,
		"config":  config,
	}

	c.HTML(http.StatusOK, "amazonwebapp.html", data)
}

func (h *AppHandler) getUserSpreadsheets(user, appName string) ([]string, error) {
	// List files in user's app directory
	path := []string{"home", user, "securestore", appName}

	item, err := h.handler.Storage.GetFile(path)
	if err != nil {
		// Directory doesn't exist, return default
		return []string{"default"}, nil
	}

	var sheets []string
	if data, ok := item.Data.([]interface{}); ok {
		for _, file := range data {
			if str, ok := file.(string); ok {
				sheets = append(sheets, str)
			}
		}
	}

	if len(sheets) == 0 {
		sheets = []string{"default"}
	}

	return sheets, nil
}

func (h *AppHandler) handleGenericApp(c *gin.Context, user, param1, paramCode, param2 string) {
	// Existing generic app handling logic
	data := gin.H{
		"fname":   param2,
		"user":    user,
		"storage": h.handler.Config.StorageBackend,
		"sheets":  []string{"Sheet1"},
		"dbLogin": 0,
	}

	c.HTML(http.StatusOK, "amazonwebapp.html", data)
}
