package api

import (
	"net/http"
)

func (rt *_router) Handler() http.Handler {
	// rotte pubbliche, senza wrap
	rt.router.GET("/", rt.getHelloWorld)
	rt.router.GET("/liveness", rt.liveness)
	// non richiede autenticazione
	rt.router.POST("/session", rt.doLogin)
	// per images
	rt.router.ServeFiles("/images/*filepath", http.Dir("./images"))
	// rotte protette da wrap
	// users
	rt.router.PUT("/users/me/name", rt.wrap(rt.setMyUserName))
	rt.router.GET("/users", rt.wrap(rt.searchUsers))
	rt.router.PUT("/users/me/photo", rt.wrap(rt.setMyPhoto))
	// conversations
	rt.router.GET("/conversations", rt.wrap(rt.getMyConversations))
	rt.router.POST("/conversations", rt.wrap(rt.createConversation))
	rt.router.POST("/conversations/:conversationId/messages", rt.wrap(rt.sendMessage))
	rt.router.GET("/conversations/:conversationId/messages", rt.wrap(rt.getConversation))
	// groups
	rt.router.POST("/groups", rt.wrap(rt.createGroup))
	rt.router.PUT("/groups/:groupId/name", rt.wrap(rt.setGroupName))
	rt.router.PUT("/groups/:groupId/photo", rt.wrap(rt.setGroupPhoto))
	rt.router.POST("/groups/:groupId/members", rt.wrap(rt.addToGroup))
	rt.router.DELETE("/groups/:groupId/members/me", rt.wrap(rt.leaveGroup))
	rt.router.GET("/groups/:groupId/members", rt.wrap(rt.getGroupMembers))

	// messages
	rt.router.POST("/messages/:messageId", rt.wrap(rt.forwardMessage))
	rt.router.DELETE("/messages/:messageId", rt.wrap(rt.deleteMessage))
	rt.router.POST("/messages/:messageId/reaction", rt.wrap(rt.commentMessage))
	rt.router.DELETE("/messages/:messageId/reaction", rt.wrap(rt.uncommentMessage))

	return rt.router
}
