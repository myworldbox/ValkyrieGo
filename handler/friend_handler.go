package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sentrionic/valkyrie/model"
	"github.com/sentrionic/valkyrie/model/apperrors"
	"log"
	"net/http"
)

/*
 * FriendHandler contains all routes related to friend actions (/api/account)
 */

// GetUserFriends returns the current users friends
func (h *Handler) GetUserFriends(c *gin.Context) {
	userId := c.MustGet("userId").(string)

	friends, err := h.friendService.GetFriends(userId)

	if err != nil {
		log.Printf("Unable to find friends for id: %v\n%v", userId, err)
		e := apperrors.NewNotFound("user", userId)

		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	c.JSON(http.StatusOK, friends)
}

// GetUserRequests returns the current users friend requests
func (h *Handler) GetUserRequests(c *gin.Context) {
	userId := c.MustGet("userId").(string)

	requests, err := h.friendService.GetRequests(userId)

	if err != nil {
		log.Printf("Unable to find requests for id: %v\n%v", userId, err)
		e := apperrors.NewNotFound("user", userId)

		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	c.JSON(http.StatusOK, requests)
}

// SendFriendRequest sends a friend request to the given member param
func (h *Handler) SendFriendRequest(c *gin.Context) {

	userId := c.MustGet("userId").(string)
	memberId := c.Param("memberId")

	if userId == memberId {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You cannot add yourself",
		})
		return
	}

	authUser, err := h.friendService.GetMemberById(userId)

	if err != nil {
		log.Printf("Unable to find user: %v\n%v", userId, err)
		e := apperrors.NewNotFound("user", userId)

		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	member, err := h.friendService.GetMemberById(memberId)

	if err != nil {
		log.Printf("Unable to find user: %v\n%v", memberId, err)
		e := apperrors.NewNotFound("user", memberId)

		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	// Check if they are already friends and no request exists
	if !isFriend(authUser, member.ID) && !containsRequest(authUser, member) {
		authUser.Requests = append(authUser.Requests, *member)
		err = h.friendService.SaveRequests(authUser)

		if err != nil {
			log.Printf("Unable to add user as friend: %v\n%v", memberId, err)
			e := apperrors.NewBadRequest("Unable to add user as friend")

			c.JSON(e.Status(), gin.H{
				"error": e,
			})
			return
		}

		// Emit friends request to the added user
		h.socketService.EmitAddFriendRequest(memberId, &model.FriendRequest{
			Id:       authUser.ID,
			Username: authUser.Username,
			Image:    authUser.Image,
			Type:     1,
		})
	}

	c.JSON(http.StatusOK, true)
}

// RemoveFriend removes the given member param from the current
// users friends.
func (h *Handler) RemoveFriend(c *gin.Context) {
	userId := c.MustGet("userId").(string)
	memberId := c.Param("memberId")

	if userId == memberId {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You cannot remove yourself",
		})
		return
	}

	authUser, err := h.friendService.GetMemberById(userId)

	if err != nil {
		log.Printf("Unable to find user: %v\n%v", memberId, err)
		e := apperrors.NewNotFound("user", memberId)

		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	member, err := h.friendService.GetMemberById(memberId)

	if err != nil {
		log.Printf("Unable to find user: %v\n%v", memberId, err)
		e := apperrors.NewNotFound("user", memberId)

		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	if isFriend(authUser, member.ID) {
		err := h.friendService.RemoveFriend(member.ID, authUser.ID)

		if err != nil {
			log.Printf("Unable to remove user from friends: %v\n%v", memberId, err)
			e := apperrors.NewBadRequest("Unable to remove the user")

			c.JSON(e.Status(), gin.H{
				"error": e,
			})
			return
		}

		// Emit signal to remove the person from the friends
		h.socketService.EmitRemoveFriend(userId, memberId)
	}

	c.JSON(http.StatusOK, true)
}

// AcceptFriendRequest accepts the friend request from the given member param
func (h *Handler) AcceptFriendRequest(c *gin.Context) {
	userId := c.MustGet("userId").(string)
	memberId := c.Param("memberId")

	if userId == memberId {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You cannot accept yourself",
		})
		return
	}

	authUser, err := h.friendService.GetMemberById(userId)

	if err != nil {
		log.Printf("Unable to find user: %v\n%v", userId, err)
		e := apperrors.NewNotFound("user", userId)

		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	member, err := h.friendService.GetMemberById(memberId)

	if err != nil {
		log.Printf("Unable to find user: %v\n%v", memberId, err)
		e := apperrors.NewNotFound("user", memberId)

		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	// Check if the current user is in the members requests
	if containsRequest(member, authUser) {
		// Add each other to friends
		authUser.Friends = append(authUser.Friends, *member)
		member.Friends = append(member.Friends, *authUser)
		err = h.friendService.SaveRequests(member)
		err = h.friendService.SaveRequests(authUser)

		if err != nil {
			log.Printf("Unable to accept friends request from user: %v\n%v", memberId, err)
			e := apperrors.NewBadRequest("Unable to accept the request")

			c.JSON(e.Status(), gin.H{
				"error": e,
			})
			return
		}

		err = h.friendService.DeleteRequest(authUser.ID, member.ID)

		if err != nil {
			log.Printf("Unable to remove user from friends: %v\n%v", memberId, err)
			e := apperrors.NewBadRequest("Unable to remove the user")

			c.JSON(e.Status(), gin.H{
				"error": e,
			})
			return
		}

		// Emit friend information to the accepted person
		h.socketService.EmitAddFriend(authUser, member)
	}

	c.JSON(http.StatusOK, true)
}

// CancelFriendRequest removes the given member param from the current
// users requests.
func (h *Handler) CancelFriendRequest(c *gin.Context) {
	userId := c.MustGet("userId").(string)
	memberId := c.Param("memberId")

	if userId == memberId {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "You cannot cancel yourself",
		})
		return
	}

	authUser, err := h.friendService.GetMemberById(userId)

	if err != nil {
		log.Printf("Unable to find user: %v\n%v", memberId, err)
		e := apperrors.NewNotFound("user", memberId)

		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	member, err := h.friendService.GetMemberById(memberId)

	if err != nil {
		log.Printf("Unable to find user: %v\n%v", memberId, err)
		e := apperrors.NewNotFound("user", memberId)

		c.JSON(e.Status(), gin.H{
			"error": e,
		})
		return
	}

	// Check if the member is in the current user's requests
	if containsRequest(authUser, member) {
		err := h.friendService.DeleteRequest(member.ID, authUser.ID)

		if err != nil {
			log.Printf("Unable to remove user from friends: %v\n%v", memberId, err)
			e := apperrors.NewBadRequest("Unable to remove the user")

			c.JSON(e.Status(), gin.H{
				"error": e,
			})
			return
		}
	}

	c.JSON(http.StatusOK, true)
}

// isFriend checks if the given users are friends
func isFriend(user *model.User, userId string) bool {
	for _, v := range user.Friends {
		if v.ID == userId {
			return true
		}
	}
	return false
}

// containsRequest checks if the given user has a friends request from the current one
func containsRequest(user *model.User, current *model.User) bool {
	for _, v := range user.Requests {
		if v.ID == current.ID {
			return true
		}
	}
	return false
}
