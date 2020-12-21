package proto

type LoginReq struct {
	RId      int    `json:"rid"`
	NickName string `json:"nickName"`
	Token    string `json:"token"`
}

type LoginRsp struct {
	RId      int    `json:"rid"`
	NickName string `json:"nickName"`
}

type LogoutReq struct {
	RId      int	`json:"RId"`
}

type LogoutRsp struct {
	RId      int	`json:"RId"`
}

type ChatReq struct {
	Type	int8	`json:"type"`	//0世界聊天
	Msg		string 	`json:"msg"`
}


type ChatMsg struct {
	RId      int    `json:"rid"`
	NickName string `json:"nickName"`
	Type	int8	`json:"type"`	//0世界聊天
	Msg		string 	`json:"msg"`
	Time	int64	`json:"time"`
}


type HistoryReq struct {
	Type	int8	`json:"type"`	//0世界聊天
}

type HistoryRsp struct {
	Type	int8		`json:"type"`	//0世界聊天
	Msgs	[]ChatMsg 	`json:"msgs"`
}