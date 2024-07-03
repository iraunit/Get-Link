package bean

type WhatsAppBusinessMessage struct {
	Object string                  `json:"object"`
	Entry  []WhatsAppBusinessEntry `json:"entry"`
}

type WhatsAppBusinessEntry struct {
	ID      string                   `json:"id"`
	Changes []WhatsAppBusinessChange `json:"changes"`
}

type WhatsAppBusinessChange struct {
	Value WhatsAppBusinessValue `json:"value"`
	Field string                `json:"field"`
}

type WhatsAppBusinessValue struct {
	MessagingProduct string                        `json:"messaging_product"`
	Metadata         WhatsAppBusinessMetadata      `json:"metadata,omitempty"`
	Contacts         []WhatsAppBusinessContact     `json:"contacts,omitempty"`
	Messages         []WhatsAppBusinessMessageData `json:"messages,omitempty"`
}

type WhatsAppBusinessMetadata struct {
	DisplayPhoneNumber string `json:"display_phone_number,omitempty"`
	PhoneNumberID      string `json:"phone_number_id,omitempty"`
}

type WhatsAppBusinessContact struct {
	Profile WhatsAppBusinessProfile `json:"profile,omitempty"`
	WaID    string                  `json:"wa_id,omitempty"`
}

type WhatsAppBusinessProfile struct {
	Name string `json:"name,omitempty"`
}

type WhatsAppBusinessMessageData struct {
	From      string                       `json:"from"`
	ID        string                       `json:"id"`
	Timestamp string                       `json:"timestamp"`
	Text      WhatsAppBusinessTextData     `json:"text,omitempty"`
	Image     WhatsAppBusinessImageData    `json:"image,omitempty"`
	Document  WhatsAppBusinessDocumentData `json:"document,omitempty"`
	Video     WhatsAppBusinessVideoData    `json:"video,omitempty"`
	Type      string                       `json:"type"`
}

type WhatsAppBusinessTextData struct {
	Body string `json:"body,omitempty"`
}

type WhatsAppBusinessImageData struct {
	MimeType string `json:"mime_type,omitempty"`
	Sha256   string `json:"sha256,omitempty"`
	ID       string `json:"id,omitempty"`
}

type WhatsAppBusinessVideoData struct {
	MimeType string `json:"mime_type,omitempty"`
	Sha256   string `json:"sha256,omitempty"`
	ID       string `json:"id,omitempty"`
}

type WhatsAppBusinessDocumentData struct {
	Filename string `json:"filename"`
	MimeType string `json:"mime_type"`
	Sha256   string `json:"sha256"`
	ID       string `json:"id"`
}

type WhatsAppBusinessSendTextMessage struct {
	MessagingProduct string                   `json:"messaging_product"`
	To               string                   `json:"to"`
	Type             string                   `json:"type"`
	Text             WhatsAppBusinessTextData `json:"text"`
}

type WhatsappMedia struct {
	Url              string `json:"url"`
	MimeType         string `json:"mime_type"`
	SHA256           string `json:"sha256"`
	FileSize         int64  `json:"file_size"`
	Id               string `json:"id"`
	MessagingProduct string `json:"messaging_product"`
}
