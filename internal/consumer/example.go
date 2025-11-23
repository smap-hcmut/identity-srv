package consumer

// import (
// 	eventConsumer "gitlab.com/gma-vietnam/tanca-connect/internal/event/delivery/rabbitmq/consumer"
// 	eventMongo "gitlab.com/gma-vietnam/tanca-connect/internal/event/repository/mongo"
// 	eventUC "gitlab.com/gma-vietnam/tanca-connect/internal/event/usecase"

// 	postConsumer "gitlab.com/gma-vietnam/tanca-connect/internal/post/delivery/rabbitmq/consumer"
// 	postProd "gitlab.com/gma-vietnam/tanca-connect/internal/post/delivery/rabbitmq/producer"
// 	postMongo "gitlab.com/gma-vietnam/tanca-connect/internal/post/repository/mongo"
// 	postUC "gitlab.com/gma-vietnam/tanca-connect/internal/post/usecase"

// 	fileConsumer "gitlab.com/gma-vietnam/tanca-connect/internal/file/delivery/rabbitmq/consumer"
// 	fileProd "gitlab.com/gma-vietnam/tanca-connect/internal/file/delivery/rabbitmq/producer"
// 	fileMongo "gitlab.com/gma-vietnam/tanca-connect/internal/file/repository/mongo"
// 	fileUC "gitlab.com/gma-vietnam/tanca-connect/internal/file/usecase"

// 	folderMongo "gitlab.com/gma-vietnam/tanca-connect/internal/folder/repository/mongo"
// 	folderUC "gitlab.com/gma-vietnam/tanca-connect/internal/folder/usecase"

// 	settingMongo "gitlab.com/gma-vietnam/tanca-connect/internal/settings/repository/mongo"
// 	settingUC "gitlab.com/gma-vietnam/tanca-connect/internal/settings/usecase"

// 	elementMongo "gitlab.com/gma-vietnam/tanca-connect/internal/element/repository/mongo"
// 	elementUC "gitlab.com/gma-vietnam/tanca-connect/internal/element/usecase"

// 	deviceMongo "gitlab.com/gma-vietnam/tanca-connect/internal/device/repository/mongo"
// 	deviceUC "gitlab.com/gma-vietnam/tanca-connect/internal/device/usecase"

// 	roomMongo "gitlab.com/gma-vietnam/tanca-connect/internal/room/repository/mongo"
// 	roomUC "gitlab.com/gma-vietnam/tanca-connect/internal/room/usecase"

// 	shopSrv "gitlab.com/gma-vietnam/tanca-connect/pkg/microservice/shop"

// 	eventProd "gitlab.com/gma-vietnam/tanca-connect/internal/event/delivery/rabbitmq/producer"
// 	eventCategoryMongo "gitlab.com/gma-vietnam/tanca-connect/internal/eventcategory/repository/mongo"
// 	eventCategoryUC "gitlab.com/gma-vietnam/tanca-connect/internal/eventcategory/usecase"

// 	rejectReasonMongo "gitlab.com/gma-vietnam/tanca-connect/internal/rejectreason/repository/mongo"
// 	rejectReasonUC "gitlab.com/gma-vietnam/tanca-connect/internal/rejectreason/usecase"

// 	commentSrv "gitlab.com/gma-vietnam/tanca-connect/pkg/microservice/comment"
// 	mediaSrv "gitlab.com/gma-vietnam/tanca-connect/pkg/microservice/media"
// )

// // Run runs the consumer server
// func (s Server) Run() error {
// 	forever := make(chan bool)

// 	// microsevice
// 	shopSrv := shopSrv.New(s.l, s.microservice.TANCA_SHOP, s.internalKey)
// 	mediaSrv := mediaSrv.New(s.l, s.microservice.TANCA_MEDIA)
// 	commentSrv := commentSrv.New(s.l, s.microservice.TANCA_COMMENT, s.internalKey)

// 	// Producer
// 	eventProd := eventProd.New(s.l, s.conn)
// 	if err := eventProd.Run(); err != nil {
// 		return err
// 	}

// 	postProd := postProd.New(s.l, s.conn)
// 	if err := postProd.Run(); err != nil {
// 		return err
// 	}

// 	fileProd := fileProd.New(s.l, s.conn)
// 	if err := fileProd.Run(); err != nil {
// 		return err
// 	}

// 	// Repositories
// 	eventCategoryRepo := eventCategoryMongo.New(s.l, s.db)
// 	elementRepo := elementMongo.New(s.l, s.db)
// 	deviceRepo := deviceMongo.New(s.l, s.db)
// 	roomRepo := roomMongo.New(s.l, s.db)
// 	eventRepo := eventMongo.New(s.l, s.db)
// 	postRepo := postMongo.New(s.l, s.db)
// 	settingRepo := settingMongo.New(s.l, s.db)
// 	rejectReasonRepo := rejectReasonMongo.New(s.l, s.db)
// 	fileRepo := fileMongo.New(s.l, s.db)
// 	folderRepo := folderMongo.New(s.l, s.db, s.redis)

// 	// Usecases
// 	elementUC := elementUC.New(s.l, elementRepo)
// 	eventCategoryUC := eventCategoryUC.New(s.l, eventCategoryRepo, elementUC)
// 	deviceUC := deviceUC.New(s.l, deviceRepo, elementUC)
// 	roomUC := roomUC.New(s.l, roomRepo, deviceUC, elementUC, nil, shopSrv)
// 	eventUC := eventUC.New(s.l, eventRepo, deviceUC, elementUC, roomUC, shopSrv, eventCategoryUC, eventProd)
// 	settingUC := settingUC.New(s.l, settingRepo, shopSrv)
// 	rejectReasonUC := rejectReasonUC.New(s.l, rejectReasonRepo, shopSrv)
// 	postUC := postUC.New(s.l, postRepo, shopSrv, settingUC, mediaSrv, commentSrv, rejectReasonUC, postProd)
// 	folderUC := folderUC.New(s.l, folderRepo)
// 	fileUC := fileUC.New(s.l, fileRepo, shopSrv, mediaSrv, folderUC, fileProd)

// 	// Set usecase
// 	roomUC.SetEventUseCase(eventUC)

// 	// Start the consumer
// 	go eventConsumer.NewConsumer(s.l, &s.conn, eventUC).Consume()
// 	go postConsumer.NewConsumer(s.l, &s.conn, postUC).Consume()
// 	go fileConsumer.NewConsumer(s.l, &s.conn, fileUC).Consume()

// 	// Keep the program running
// 	<-forever

// 	return nil
// }
