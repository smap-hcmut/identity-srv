package consumer

import (
	"github.com/nguyentantai21042004/smap-api/config"
	"github.com/nguyentantai21042004/smap-api/internal/appconfig/mongo"
	"github.com/nguyentantai21042004/smap-api/internal/appconfig/oauth"
	"github.com/nguyentantai21042004/smap-api/internal/consumer"
	pkgCrt "github.com/nguyentantai21042004/smap-api/pkg/encrypter"
	pkgLog "github.com/nguyentantai21042004/smap-api/pkg/log"
	"github.com/nguyentantai21042004/smap-api/pkg/rabbitmq"
	pkgRedis "github.com/nguyentantai21042004/smap-api/pkg/redis"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	crp := pkgCrt.NewEncrypter(cfg.Encrypter.Key)

	client, err := mongo.Connect(cfg.Mongo, crp)
	if err != nil {
		panic(err)
	}
	db := client.Database(cfg.Mongo.Database)
	defer mongo.Disconnect(client)

	conn, err := rabbitmq.Dial(cfg.RabbitMQConfig.URL, true)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	redisClient, err := pkgRedis.Connect(pkgRedis.NewClientOptions().SetOptions(cfg.Redis))
	if err != nil {
		panic(err)
	}
	defer redisClient.Disconnect()

	oauthConfig := oauth.NewOauthConfig(cfg.Oauth)

	l := pkgLog.InitializeZapLogger(pkgLog.ZapConfig{
		Level:    cfg.Logger.Level,
		Mode:     cfg.Logger.Mode,
		Encoding: cfg.Logger.Encoding,
	})

	srv, err := consumer.New(l, consumer.ConsumerConfig{
		Encrypter:    crp,
		JwtSecretKey: cfg.JWT.SecretKey,
		InternalKey:  cfg.InternalConfig.InternalKey,
		MongoDB:      &db,
		SMTPConfig:   cfg.SMTP,
		AMQPConn:     conn,
		RedisClient:  &redisClient,
		OauthConfig:  oauthConfig,
	})
	if err != nil {
		panic(err)
	}

	err = srv.Run()
	if err != nil {
		panic(err)
	}

}
