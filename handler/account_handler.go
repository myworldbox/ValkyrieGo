package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sentrionic/valkyrie/model/apperrors"
	"log"
	"mime/multipart"
	"net/http"
)

/*
 * AccountHandler contains all routes related to account actions (/api/account)
 * that the authenticated user can do
 */

// Me handler calls services for getting
// a user's details
// Me godoc
// @Tags Account
// @Summary Get Current User
// @Produce  json
// @Success 200 {object} model.User
// @Router /account [get]
func (h *Handler) Me(c *gin.Context) {
	userId := c.MustGet("userId").(string)
	user, err := h.userService.Get(userId)

	if err != nil {
		log.Printf("Unable to find user: %v\n%v", userId, err)
		e := apperrors.NewNotFound("user", userId)

		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

type editReq struct {
	// Min 3, max 30 characters.
	Username string `form:"username" binding:"required,min=3,max=30"`
	// Must be unique
	Email string `form:"email" binding:"required,email"`
	// image/png or image/jpeg
	Image *multipart.FileHeader `form:"image" binding:"omitempty" swaggertype:"string" format:"binary"`
} //@name EditUser

// Edit handler edits the users account details
// Edit godoc
// @Tags Account
// @Summary Update Current User
// @Accept mpfd
// @Produce  json
// @Param account body editReq true "Update Account"
// @Success 200 {object} model.User
// @Router /account [put]
func (h *Handler) Edit(c *gin.Context) {
	userId := c.MustGet("userId").(string)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.MaxBodyBytes)

	var req editReq

	if ok := bindData(c, &req); !ok {
		return
	}

	authUser, err := h.userService.Get(userId)

	if err != nil {
		err := errors.New("provided session is invalid")
		c.JSON(401, gin.H{
			"error": err,
		})
		c.Abort()
		return
	}

	authUser.Username = req.Username

	// New email, check if it's unique
	if authUser.Email != req.Email {
		inUse := h.userService.IsEmailAlreadyInUse(req.Email)

		if inUse {
			c.JSON(http.StatusBadRequest, gin.H{
				"field":   "Email",
				"message": "email already in use",
			})
			return
		}
		authUser.Email = req.Email
	}

	if req.Image != nil {

		// Validate image mime-type is allowable
		mimeType := req.Image.Header.Get("Content-Type")

		if valid := isAllowedImageType(mimeType); !valid {
			e := apperrors.NewBadRequest("imageFile must be 'image/jpeg' or 'image/png'")
			c.JSON(e.Status(), gin.H{
				"error": e,
			})
			return
		}

		directory := fmt.Sprintf("valkyrie_go/users/%s", authUser.ID)
		url, err := h.userService.ChangeAvatar(req.Image, directory)

		if err != nil {
			c.JSON(500, gin.H{
				"error": err,
			})
			return
		}

		_ = h.userService.DeleteImage(authUser.Image)

		authUser.Image = url
	}

	err = h.userService.UpdateAccount(authUser)

	if err != nil {
		log.Printf("Failed to update user: %v\n", err.Error())

		c.JSON(apperrors.Status(err), gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, authUser)
}

type changeRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	// Min 6, max 150 characters.
	NewPassword string `json:"newPassword" binding:"required,gte=6,lte=150"`
	// Must be the same as the newPassword value.
	ConfirmNewPassword string `json:"confirmNewPassword" binding:"required,gte=6,lte=150"`
} //@name ChangePasswordRequest

// ChangePassword handler changes the users password
// ChangePassword godoc
// @Tags Account
// @Summary Change Current User's Password
// @Accept json
// @Produce  json
// @Param request body changeRequest true "Change Password"
// @Success 200 {object} model.Success
// @Router /account/change-password [put]
func (h *Handler) ChangePassword(c *gin.Context) {
	userId := c.MustGet("userId").(string)
	var req changeRequest

	// Bind incoming json to struct and check for validation errors
	if ok := bindData(c, &req); !ok {
		return
	}

	// Check if passwords are equal
	if req.NewPassword != req.ConfirmNewPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"field":   "password",
			"message": "passwords do not match",
		})
		return
	}

	authUser, err := h.userService.Get(userId)

	if err != nil {
		err := errors.New("provided session is invalid")
		c.JSON(401, gin.H{
			"error": err,
		})
		c.Abort()
		return
	}

	err = h.userService.ChangePassword(req.CurrentPassword, req.NewPassword, authUser)

	if err != nil {
		log.Printf("Failed to change password: %v\n", err.Error())
		c.JSON(apperrors.Status(err), gin.H{
			"error": err,
		})
		return
	}

	c.JSON(http.StatusOK, true)
}
