package routes

import (
	"RyanForce/controllers"
	"RyanForce/middleware"
	"RyanForce/utils"
	"RyanForce/web"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouterWithEngine initializes API routes and WebUI routes
func SetupRouterWithEngine(r *gin.Engine) *gin.Engine {

	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/static", "./web/static")

	// Root route
	r.GET("/", func(c *gin.Context) {
		token, err := c.Cookie("token")
		if err == nil {
			if _, err := utils.ParseJWT(token); err == nil {
				c.Redirect(http.StatusFound, "/dashboard")
				return
			}
		}
		c.Redirect(http.StatusFound, "/login")
	})

	// Authentication and Session
	r.GET("/login", web.ShowLoginPage)
	r.POST("/login", web.HandleWebLogin)
	r.GET("/dashboard", web.ShowDashboard)
	r.GET("/logout", web.HandleLogout)

	r.GET("/reset-password", web.ShowResetForm)
	r.POST("/reset-password", web.HandleResetPassword)

	r.GET("/admin/reset-password", web.ShowAdminResetForm)
	r.POST("/admin/reset-password", web.HandleAdminResetPassword)

	r.GET("/admin/unlock", web.ShowUnlockForm)
	r.POST("/admin/unlock", web.HandleUnlockUser)

	r.GET("/admin/unassigned-tickets", web.ShowUnassignedTickets)

	// Group: Ticket WebUI - Protected
	ticketGroup := r.Group("/tickets")
	ticketGroup.Use(middleware.WebAuthMiddleware())
	{
		ticketGroup.GET("/create", web.ShowCreateTicketForm)
		ticketGroup.POST("/create", web.HandleCreateTicket)
		ticketGroup.GET("/mine", web.ListClientTickets)
		ticketGroup.GET("/tech", web.ListTechTickets) // << ADD BACK
		ticketGroup.GET("/:id", web.ViewTicketPage)
		ticketGroup.POST("/:id/comments", web.AddComment)
		ticketGroup.GET("/update/:id", web.ShowUpdateTicketForm)
		ticketGroup.POST("/update/:id", web.HandleUpdateTicket)
	}

	// Group: Comment WebUI - Protected
	commentGroup := r.Group("/comments")
	commentGroup.Use(middleware.WebAuthMiddleware())
	{
		commentGroup.GET("/:commentID/edit", web.ShowEditCommentForm)
		commentGroup.POST("/:commentID/update", web.UpdateComment)
		commentGroup.POST("/:commentID/delete", web.DeleteComment)
	}

	// Admin
	adminGroup := r.Group("/admin")
	{
		adminGroup.GET("/tickets/:id/assign", controllers.FindMatchingTechs)
		adminGroup.POST("/tickets/:id/assign/:tech_id", controllers.AssignTechToTicket)
		adminGroup.GET("/assigned-tickets", web.ShowAssignedTickets)
		adminGroup.POST("/tickets/:id/unassign", controllers.UnassignTechFromTicket)

		adminGroup.GET("/clients", web.ListClients)
		adminGroup.GET("/clients/new", web.NewClientForm)
		adminGroup.POST("/clients", web.CreateClient)
		adminGroup.GET("/clients/:id", web.ShowClient)
		adminGroup.GET("/clients/:id/edit", web.ShowEditClientForm)
		adminGroup.POST("/clients/:id", web.UpdateClient)
		adminGroup.POST("/clients/:id/delete", web.DeleteClient)

		adminGroup.GET("/techs", web.ListTechs)
		adminGroup.GET("/techs/new", web.NewTechForm)
		adminGroup.POST("/techs", web.CreateTech)
		adminGroup.GET("/techs/:id", web.ShowTech)
		adminGroup.GET("/techs/:id/edit", web.EditTechForm)
		adminGroup.POST("/techs/:id", web.UpdateTech)
		adminGroup.POST("/techs/:id/delete", web.DeleteTech)

		adminGroup.GET("/accounts", web.ListAccounts)
		adminGroup.POST("/accounts", web.CreateAccount)
		adminGroup.GET("/accounts/:id/edit", web.EditAccountForm)
		adminGroup.POST("/accounts/:id", web.UpdateAccount)
		adminGroup.POST("/accounts/:id/delete", web.DeleteAccount)

		adminGroup.GET("/reports", web.AdminReports)
		adminGroup.GET("/reports/export", web.ExportReportCSV)
		adminGroup.GET("/clients/export", web.ExportClientsCSV)
		adminGroup.GET("/reports/audit/export", web.ExportAuditCSV)
	}

	// REST API
	r.POST("/api/login", controllers.LoginAPI)
	r.POST("/api/register", controllers.RegisterAPI)

	protected := r.Group("/api")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		protected.GET("/tickets", controllers.ListTicketsAPI)
		protected.GET("/tickets/filter", controllers.FilterTicketsAPI)
		protected.POST("/tickets", controllers.CreateTicketAPI)
		protected.GET("/tickets/:id", controllers.ViewTicketAPI)
		protected.PATCH("/tickets/:id", controllers.UpdateTicketAPI)
		protected.DELETE("/tickets/:id", controllers.DeleteTicketAPI)
		protected.POST("/tickets/:id/assign", controllers.AssignTicketAPI)

		protected.GET("/users", controllers.GetUsers)
		protected.DELETE("/users/:id", controllers.DeleteUserAPI)
	}

	// 404 fallback
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", nil)
	})

	return r
}
