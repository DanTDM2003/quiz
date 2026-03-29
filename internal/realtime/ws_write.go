package realtime

import "github.com/gorilla/websocket"

func WriteWelcome(c *websocket.Conn, quizID, participantID string) error {
	return c.WriteJSON(welcomeMsg{
		Type:          "welcome",
		QuizID:        quizID,
		ParticipantID: participantID,
	})
}

func WriteError(c *websocket.Conn, code, msg string) error {
	return c.WriteJSON(wsErrMsg{
		Type:    "error",
		Code:    code,
		Message: msg,
	})
}
