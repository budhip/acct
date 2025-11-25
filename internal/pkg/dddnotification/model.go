package dddnotification

const SERVICE_NAME string = "ddd-notification"

type (
	RequestEmail struct {
		From        string        `json:"from"`
		FromName    string        `json:"fromName"`
		To          string        `json:"to"`
		ToName      string        `json:"toName"`
		BCC         string        `json:"bcc"`
		BCCname     string        `json:"bccName"`
		CC          []Cc          `json:"cc"`
		ReplyTo     string        `json:"replyTo"`
		Template    string        `json:"template"`
		Subject     string        `json:"subject"`
		Bucket      bool          `json:"bucket"`
		Subs        []interface{} `json:"subs"`
		Attachments []Attachment  `json:"attachments"`
	}

	ParamSubscribe struct {
		Email  string `json:"email"`
		Name   string `json:"name"`
		Status string `json:"status"`
		NoSubs string
	}
	Attachment struct {
		Type    string `json:"type"`
		Name    string `json:"name"`
		Content string `json:"content"`
	}
	Cc struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
)

type (
	PayloadNotification struct {
		Title        string      `json:"title"`
		Service      string      `json:"service"`
		SlackChannel string      `json:"slackChannel"`
		Data         MessageData `json:"data"`
	}
	MessageData struct {
		Operation string `json:"Operation"`
		Message   string `json:"Message"`
	}
)

type ResponseSendMessage struct {
	Status  int                    `json:"status"`
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}
