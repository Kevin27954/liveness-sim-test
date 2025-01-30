package node

type Hub struct {
	connList []string
}

func (h *Hub) sendMessage(message string) {
}

func (h *Hub) addConn(conn string) {
	h.connList = append(h.connList, conn)
}
