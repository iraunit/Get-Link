package util

const (
	PRODUCTION                      = "production"
	DEVELOPMENT                     = "development"
	WhatsappCloudApiSendMessage     = `https://graph.facebook.com/v19.0/%s/messages`
	WhatsappCloudApiGetMediaDataUrl = `https://graph.facebook.com/v20.0/%s`
	ChromeExtensionUrl              = "chrome-extension://pcphjmlofajahcidbgfgphicmmdfkdif"
	PathToFiles                     = "/tmp/data/%s/%s"
	FreeWhatsappFileLimitSizeMB     = 100
	PremiumWhatsappFileLimitSizeMB  = 200
)

// routes

const (
	WhatsappWebhook     = "/whatsapp-webhook"
	VerifyWhatsappEmail = "/verify-whatsapp-email"
)

const (
	UUID          = "uuid"
	EMAIL         = "email"
	DEVICE        = "device"
	AUTHORIZATION = "Authorization"
	WHATSAPP      = "whatsapp"
	TELEGRAM      = "telegram"
	GETLINK       = "getlink"
)
