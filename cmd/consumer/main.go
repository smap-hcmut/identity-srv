package main

import (
	
)

// func main() {
// 	ctx := context.Background()

// 	// Load config
// 	cfg, err := config.Load()
// 	if err != nil {
// 		panic(err)
// 	}

// 	i18n.InitI18n()

// 	l := pkgLog.InitializeZapLogger(pkgLog.ZapConfig{
// 		Level:    cfg.Logger.Level,
// 		Mode:     cfg.Logger.Mode,
// 		Encoding: cfg.Logger.Encoding,
// 	})

// 	crp := pkgCrt.NewEncrypter(cfg.Encrypter.Key)

// 	client, err := mongo.Connect(cfg.Mongo, crp)
// 	if err != nil {
// 		l.Fatalf(ctx, "Failed to connect to MongoDB: %v", err)
// 	}
// 	defer mongo.Disconnect(client)

// 	db := client.Database(cfg.Mongo.Database)

// 	conn, err := rabbitmq.Dial(cfg.RabbitMQConfig.URL, true)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer conn.Close()

// 	redisClient, err := redis.Connect(cfg.RedisConfig)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer redisClient.Disconnect()

// 	microservice := consumer.TancaMicroserviceConfig{
// 		TANCA_TIMESHEET:    cfg.TancaMicroservice.TANCA_TIMESHEET,
// 		TANCA_AUTH:         cfg.TancaMicroservice.TANCA_AUTH,
// 		TANCA_SHOP:         cfg.TancaMicroservice.TANCA_SHOP,
// 		TANCA_NOTIFICATION: cfg.TancaMicroservice.TANCA_NOTIFICATION,
// 		TANCA_REQUEST:      cfg.TancaMicroservice.TANCA_REQUEST,
// 		TANCA_TASK:         cfg.TancaMicroservice.TANCA_TASK,
// 		TANCA_INTEGRATION:  cfg.TancaMicroservice.TANCA_INTEGRATION,
// 	}

// 	config := consumer.ServerConfig{
// 		Conn:              conn,
// 		DB:                db,
// 		Redis:             redisClient,
// 		TancaMicroservice: microservice,
// 		Encrypter:         crp,
// 		InternalKey:       cfg.InternalConfig.InternalKey,
// 	}

// 	if err := consumer.NewServer(l, config).Run(); err != nil {
// 		l.Fatalf(ctx, "Failed to run consumer server: %v", err)
// 	}
// }
