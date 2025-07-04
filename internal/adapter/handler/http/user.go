package http

import (
	"harajuku/backend/internal/core/domain"
	"harajuku/backend/internal/core/port"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UserHandler represents the HTTP handler for user-related requests
type UserHandler struct {
	svc port.UserService
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(svc port.UserService) *UserHandler {
	return &UserHandler{
		svc,
	}
}

// registerRequest represents the request body for creating a user
type registerRequest struct {
	Name           string `json:"name" binding:"required" example:"John"`
	LastName       string `json:"lastName" binding:"required" example:"Doe"`
	SecondLastName string `json:"SecondLastName" example:"Doe"`
	Email          string `json:"email" binding:"required,email" example:"test@example.com"`
	Password       string `json:"password" binding:"required,min=8" example:"12345678"`
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	create a new user account with default role "cashier"
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			registerRequest	body		registerRequest	true	"Register request"
//	@Success		200				{object}	userResponse	"User created"
//	@Failure		400				{object}	errorResponse	"Validation error"
//	@Failure		401				{object}	errorResponse	"Unauthorized error"
//	@Failure		404				{object}	errorResponse	"Data not found error"
//	@Failure		409				{object}	errorResponse	"Data conflict error"
//	@Failure		500				{object}	errorResponse	"Internal server error"
//	@Router			/users [post]
func (uh *UserHandler) Register(ctx *gin.Context) {
	var req registerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validationError(ctx, err)
		return
	}

	user_id := uuid.New()

	user := domain.User{
		ID:             user_id,
		Name:           req.Name,
		Email:          req.Email,
		LastName:       req.LastName,
		SecondLastName: req.SecondLastName,
		Password:       req.Password,
	}

	_, err := uh.svc.Register(ctx, &user)
	if err != nil {
		handleError(ctx, err)
		return
	}

	rsp := newUserResponse(&user)

	handleSuccess(ctx, rsp)
}

// listUsersRequest represents the request body for listing users
type listUsersRequest struct {
	Skip    uint64             `form:"skip" binding:"min=0" example:"0"`
	Limit   uint64             `form:"limit" binding:"required,min=1" example:"5"`
	Filters userFiltersRequest `form:"filters"`
}

// userFiltersRequest represents the filter criteria for listing users
type userFiltersRequest struct {
	Name           string `form:"name" example:"John"`
	LastName       string `form:"lastName" example:"Doe"`
	SecondLastName string `form:"secondLastName" example:"Smith"`
	Role           string `form:"role" validate:"omitempty,oneof=admin client" example:"client" enums:"admin,client"`
}

// ListUsers godoc
//
//	@Summary		List users
//	@Description	List users with pagination and filtering
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			skip		query		uint64				false	"Skip"	default(0)
//	@Param			limit		query		uint64				false	"Limit"	default(10)
//	@Param			filters		query		userFiltersRequest	false	"Filters"
//	@Success		200			{object}	meta				"Users displayed"
//	@Failure		400			{object}	errorResponse		"Validation error"
//	@Failure		500			{object}	errorResponse		"Internal server error"
//	@Router			/users [get]
//	@Security		BearerAuth
func (uh *UserHandler) ListUsers(ctx *gin.Context) {
	var req listUsersRequest
	var usersList []userResponse

	// First bind the basic parameters
	if err := ctx.ShouldBindQuery(&req); err != nil {
		validationError(ctx, err)
		return
	}

	// Manually extract filter parameters
	filters := domain.UserFilters{
		Name:           ctx.Query("filters.name"),
		LastName:       ctx.Query("filters.lastName"),
		SecondLastName: ctx.Query("filters.secondLastName"),
		Role:           domain.UserRole(ctx.Query("filters.role")),
	}

	// Debug output
	slog.Info("Filter parameters",
		"name", filters.Name,
		"lastName", filters.LastName,
		"role", filters.Role,
		"rawQuery", ctx.Request.URL.RawQuery,
	)

	users, err := uh.svc.ListUsers(ctx, req.Skip, req.Limit, filters)
	if err != nil {
		handleError(ctx, err)
		return
	}

	for _, user := range users {
		usersList = append(usersList, newUserResponse(&user))
	}

	total := uint64(len(usersList))
	meta := newMeta(total, req.Limit, req.Skip)
	rsp := toMap(meta, usersList, "users")

	handleSuccess(ctx, rsp)
}

// getUserRequest represents the request body for getting a user
type getUserRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,min=1" example:"1"`
}

// GetUser godoc
//
//	@Summary		Get a user
//	@Description	Get a user by id
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64			true	"User ID"
//	@Success		200	{object}	userResponse	"User displayed"
//	@Failure		400	{object}	errorResponse	"Validation error"
//	@Failure		404	{object}	errorResponse	"Data not found error"
//	@Failure		500	{object}	errorResponse	"Internal server error"
//	@Router			/users/{id} [get]
//	@Security		BearerAuth
func (uh *UserHandler) GetUser(ctx *gin.Context) {
	var req getUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		validationError(ctx, err)
		return
	}

	user, err := uh.svc.GetUser(ctx, req.ID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	rsp := newUserResponse(user)

	handleSuccess(ctx, rsp)
}

// updateUserRequest represents the request body for updating a user
type updateUserRequest struct {
	Name           string          `json:"name" binding:"omitempty,required" example:"John Doe"`
	LastName       string          `json:"name" binding:"omitempty,required" example:"John Doe"`
	SecondLastName string          `json:"name" binding:"omitempty,required" example:"John Doe"`
	Email          string          `json:"email" binding:"omitempty,required,email" example:"test@example.com"`
	Password       string          `json:"password" binding:"omitempty,required,min=8" example:"12345678"`
	Role           domain.UserRole `json:"role" binding:"omitempty,required,user_role" example:"admin"`
}

// UpdateUser godoc
//
//	@Summary		Update a user
//	@Description	Update a user's name, email, password, or role by id
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			id					path		uint64				true	"User ID"
//	@Param			updateUserRequest	body		updateUserRequest	true	"Update user request"
//	@Success		200					{object}	userResponse		"User updated"
//	@Failure		400					{object}	errorResponse		"Validation error"
//	@Failure		401					{object}	errorResponse		"Unauthorized error"
//	@Failure		403					{object}	errorResponse		"Forbidden error"
//	@Failure		404					{object}	errorResponse		"Data not found error"
//	@Failure		500					{object}	errorResponse		"Internal server error"
//	@Router			/users/{id} [put]
//	@Security		BearerAuth
func (uh *UserHandler) UpdateUser(ctx *gin.Context) {
	var req updateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		validationError(ctx, err)
		return
	}

	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)

	if err != nil {
		validationError(ctx, err)
		return
	}

	user := domain.User{
		ID:             id,
		Name:           req.Name,
		LastName:       req.LastName,
		SecondLastName: req.SecondLastName,
		Email:          req.Email,
		Password:       req.Password,
		Role:           req.Role,
	}

	_, err = uh.svc.UpdateUser(ctx, &user)
	if err != nil {
		handleError(ctx, err)
		return
	}

	rsp := newUserResponse(&user)

	handleSuccess(ctx, rsp)
}

// deleteUserRequest represents the request body for deleting a user
type deleteUserRequest struct {
	ID uuid.UUID `uri:"id" binding:"required,min=1" example:"1"`
}

// DeleteUser godoc
//
//	@Summary		Delete a user
//	@Description	Delete a user by id
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64			true	"User ID"
//	@Success		200	{object}	response		"User deleted"
//	@Failure		400	{object}	errorResponse	"Validation error"
//	@Failure		401	{object}	errorResponse	"Unauthorized error"
//	@Failure		403	{object}	errorResponse	"Forbidden error"
//	@Failure		404	{object}	errorResponse	"Data not found error"
//	@Failure		500	{object}	errorResponse	"Internal server error"
//	@Router			/users/{id} [delete]
//	@Security		BearerAuth
func (uh *UserHandler) DeleteUser(ctx *gin.Context) {
	var req deleteUserRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		validationError(ctx, err)
		return
	}

	err := uh.svc.DeleteUser(ctx, req.ID)
	if err != nil {
		handleError(ctx, err)
		return
	}

	handleSuccess(ctx, nil)
}
