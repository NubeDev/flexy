package appcommon

//func (app *App) InitNATS(natsURL string) error {
//	var err error
//	app.NatsConn, err = nats.Connect(natsURL)
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func (app *App) Publish(operation string, subject string, msg []byte) error {
//	topic := fmt.Sprintf("%s.%s.%s", app.AppID, operation, subject)
//	return app.NatsConn.Publish(topic, msg)
//}
//
//func (app *App) Subscribe(operation string, subject string, globalSubject bool, handler nats.MsgHandler) (*nats.Subscription, error) {
//	topic := fmt.Sprintf("%s.%s.%s", app.AppID, operation, subject)
//	if globalSubject {
//		topic = fmt.Sprintf("global.%s.%s", operation, subject)
//	}
//	return app.NatsConn.Subscribe(topic, handler)
//}
//
//func (app *App) SubscribeRespond(operation string, subject string, globalSubject bool, handler func(msg *nats.Msg) []byte, responseSubject string) (*nats.Subscription, error) {
//	topic := fmt.Sprintf("%s.%s.%s", app.AppID, operation, subject)
//	if globalSubject {
//		topic = fmt.Sprintf("global.%s.%s", operation, subject)
//	}
//	return app.NatsConn.Subscribe(topic, func(msg *nats.Msg) {
//		// Call the handler to generate a response
//		response := handler(msg)
//
//		// If there is no reply subject, publish the response to a predefined subject
//		if msg.Reply == "" {
//			err := app.NatsConn.Publish(topic+".response", response)
//			if err != nil {
//				fmt.Println("Error publishing response to subject:", err)
//			}
//		} else {
//			// Fallback: Respond if the reply subject exists
//			err := msg.Respond(response)
//			if err != nil {
//				fmt.Println("Error responding to message:", err)
//			}
//		}
//	})
//}
//
//// Request sends a request message and waits for a response
//func (app *App) Request(operation string, subject string, msg []byte, timeout time.Duration) (*nats.Msg, error) {
//	topic := fmt.Sprintf("%s.%s.%s", app.AppID, operation, subject)
//	return app.NatsConn.Request(topic, msg, timeout)
//}
