package logger

const (
	DiscordErrorOnStartup = "failed to start the bot"

	RedisFailedParseIndexDB    = "failed to parse the Redis database index"
	RedisFailedPingDB          = "failed to ping the Redis database"
	RedisFailedCloseConnection = "failed to close the Redis database connection"

	SQLiteFailedConnectionAttempt    = "failed to connect to the database; in-memory mode will be used"
	SQLiteFailedConnectionMemoryMode = "failed to connect to in-memory mode"
	SQLiteFailedConnectionNil        = "failed to establish a database connection"
	SQLiteFailedPingDB               = "failed to ping the SQLite database"
	SQLiteFailedCloseConnection      = "failed to close the SQLite database connection"

	AutoReactionFailedInvalidateCache = "failed to invalidate auto-reaction cache"
	AutoReactionFailedCache           = "failed to cache auto-reaction"
	AutoReactionFailedCacheAll        = "failed to cache auto-reactions"

	EffectFailedCache           = "failed to cache effect"
	EffectFailedCacheAll        = "failed to cache effects"
	EffectFailedInvalidateCache = "failed to invalidate effect cache"

	AutoReactionEventFailedGet       = "failed to get auto-reaction"
	AutoReactionEventFailedAdd       = "failed to add auto-reaction"
	EffectEventFailedRetrieve        = "failed to retrieve effect"
	EffectEventFailedDeleteExpired   = "failed to delete expired effect"
	EffectEventFailedGenerateMessage = "failed to transform message with Gemini"
	EffectEventFailedCreateWebhook   = "failed to create webhook"
	EffectEventFailedDeleteWebhook   = "failed to delete webhook"
	EffectEventFailedSendWebhook     = "failed to send transformed message via Discord webhook"
	EffectEventFailedDeleteMessage   = "failed to delete original message"

	ResponderFailedResponseCause    = "failed to build responder cause"
	ResponderNilMessageContext      = "cannot send interaction message context response: context is nil"
	ResponderFailedSendMessageReply = "failed to send message reply"
	ResponderNilInteractionContext  = "cannot send interaction response: context is nil"
	ResponderInvalidEmbed           = "invalid embed in response"
	ResponderFailedEditInteraction  = "failed to edit interaction reply"
	ResponderFailedSendInteraction  = "failed to send initial interaction reply"
)
